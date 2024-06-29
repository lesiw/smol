package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/jackc/pgx/v5"
	"lesiw.io/defers"
	"lesiw.io/flag"
	"lesiw.io/short/stmt"
)

var (
	flags = flag.NewSet(os.Stderr, "short [-a ALIAS] URL")
	alias = flags.String("a", "alias")

	errParse = errors.New("parse error")

	db *pgx.Conn
)

const (
	chars    = 6
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func main() {
	if err := run(); err != nil {
		if !errors.Is(err, errParse) {
			fmt.Fprintln(os.Stderr, err)
		}
		defers.Exit(1)
	}
	defers.Exit(0)
}

func run() error {
	if err := flags.Parse(os.Args[1:]...); err != nil {
		return errParse
	}
	if len(flags.Args) < 1 {
		flags.PrintError("url is required")
		return errParse
	}
	url := flags.Args[0]
	ctx, cancel := context.WithCancel(context.Background())
	defers.Add(cancel)

	var err error
	db, err = pgx.Connect(ctx, "postgres://postgres@localhost/postgres")
	if err != nil {
		return err
	}

	var domain string
	row := db.QueryRow(ctx, stmt.GetDomain)
	if err := row.Scan(&domain); err != nil {
		return fmt.Errorf("failed to get domain: %w", err)
	}

	var id string
	if *alias == "" {
		for range 3 {
			if id, err = shortid(); err != nil {
				return err
			}
			rows, err := db.Query(ctx, stmt.AddUrl, id, url)
			if err != nil {
				continue
			}
			for rows.Next() {
				if err = rows.Scan(&id); err != nil {
					return fmt.Errorf("failed to read id: %w", err)
				}
				fmt.Printf("%s/%s\n", domain, id)
			}
			break
		}
		if err != nil {
			return fmt.Errorf("failed to add url: %w", err)
		}
	} else {
		if _, err = db.Exec(ctx, stmt.SetUrl, *alias, url); err != nil {
			return fmt.Errorf("failed to set url: %w", err)
		}
		fmt.Printf("%s/%s\n", domain, *alias)
	}

	return nil
}

func shortid() (string, error) {
	id := make([]byte, chars)
	for i := range id {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		id[i] = alphabet[num.Int64()]
	}
	return string(id), nil
}
