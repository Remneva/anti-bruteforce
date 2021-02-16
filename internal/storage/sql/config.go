package sql

import (
	"context"
	"fmt"

	"github.com/Remneva/anti-bruteforce/internal/storage"
)

var _ storage.ConfigurationStorage = (*Storage)(nil)

func (s *Storage) Configs() storage.ConfigurationStorage {
	return s
}
func (s *Storage) Get(ctx context.Context) (map[string]int64, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT * FROM configurations
	`)
	if err != nil {
		return nil, fmt.Errorf("query error %w", err)
	}
	defer rows.Close()
	configs := make(map[string]int64, 3)

	for rows.Next() {
		var c storage.Config
		if err = rows.Scan(
			&c.Key,
			&c.Value,
		); err != nil {
			return nil, fmt.Errorf("query error %w", err)
		}
		configs[c.Key] = c.Value
	}
	return configs, rows.Err()
}
