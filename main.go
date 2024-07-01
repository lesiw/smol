//go:generate go run lesiw.io/plain/cmd/plaingen@v0.4.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"lesiw.io/defers"
	"lesiw.io/plain"
	"lesiw.io/smol/internal/stmt"
)

var db *pgxpool.Pool

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		defers.Exit(1)
	}
	defers.Exit(0)
}

func run() error {
	ctx := context.Background()
	db = plain.ConnectPgx(ctx)
	defers.Add(db.Close)

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return fmt.Errorf("bad DOMAIN: environment variable not set")
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
