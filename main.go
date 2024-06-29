//go:generate go run internal/sql.go

package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"lesiw.io/defers"
	"lesiw.io/smol/stmt"
)

//go:embed sql/migrations/*
var migrations embed.FS

var db *pgx.Conn

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		defers.Exit(1)
	}
	defers.Exit(0)
}

func run() error {
	ctx := context.Background()

	var err error
	for range 3 {
		db, err = pgx.Connect(ctx, os.Getenv("POSTGRES_URL"))
		if err != nil {
			time.Sleep(time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	defers.Add(func() { _ = db.Close(ctx) })

	src, err := iofs.New(migrations, "sql/migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance(
		"iofs",
		src,
		strings.Replace(os.Getenv("POSTGRES_URL"),
			"postgres://", "pgx5://", 1),
	)
	if err != nil {
		return fmt.Errorf("failed to set up migration: %w", err)
	}
	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return fmt.Errorf("DOMAIN must be set")
	}
	if _, err := db.Exec(ctx, stmt.SetDomain, domain); err != nil {
		return fmt.Errorf("failed to store domain: %w", err)
	}

	http.HandleFunc("/{id}", Redirect)
	return http.ListenAndServe(":8080", nil)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id := r.PathValue("id")
	row := db.QueryRow(ctx, stmt.GetUrl, id)
	var url string
	_ = row.Scan(&url)
	if url == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
