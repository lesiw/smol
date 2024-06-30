package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var file *os.File

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate stmts.go: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	var err error
	file, err = os.OpenFile("stmt/stmts.go",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	_, _ = file.WriteString("package stmt\n\n")
	return filepath.WalkDir("sql/statements", walkFunc)
}

func walkFunc(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	b := &strings.Builder{}
	_, _ = io.Copy(b, f)
	_, _ = file.WriteString(fmt.Sprintf(
		"const %s = `\n%s`\n\n",
		name(d.Name()),
		b.String(),
	))
	return nil
}

func name(s string) string {
	s, _, _ = strings.Cut(s, ".")
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '-' || r == '_' {
			upper = true
			continue
		}
		if upper || unicode.IsUpper(r) {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}
