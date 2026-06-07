package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	names, err := migrationNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", name, err)
		}
	}
	return nil
}

func migrationSQL() (string, error) {
	names, err := migrationNames()
	if err != nil {
		return "", err
	}

	var combined strings.Builder
	for _, name := range names {
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return "", fmt.Errorf("read migration %s: %w", name, err)
		}
		combined.Write(content)
		combined.WriteString("\n")
	}
	return combined.String(), nil
}

func migrationNames() ([]string, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
