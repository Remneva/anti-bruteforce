package sql

import (
	"context"
	"fmt"

	"github.com/Remneva/anti-bruteforce/internal/storage"
	"go.uber.org/zap"
)

var _ storage.ListStorage = (*Storage)(nil)

func (s *Storage) Lists() storage.ListStorage {
	return s
}
func (s *Storage) AddToWhiteList(ctx context.Context, ip storage.IP) error {
	query := `INSERT INTO whitelist (ip, mask)
VALUES($1, $2)`
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.DB.ExecContext(ctx, query, ip.IP, ip.Mask)
	if err != nil {
		return fmt.Errorf("query error %w", err)
	}
	return nil
}

func (s *Storage) DeleteFromWhiteList(ctx context.Context, ip storage.IP) error {
	query := `DELETE from whitelist WHERE ip = $1`
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.DB.ExecContext(ctx, query, ip.IP)
	if err != nil {
		s.l.Error("query error", zap.Error(err))
		return fmt.Errorf("open connection error %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected error %w", err)
	}
	if rowsAffected > 0 {
		s.l.Info("ip deleted from white list", zap.String("ip", ip.IP))
	} else {
		s.l.Info("ip does not exist in white list", zap.String("ip", ip.IP))
	}
	return nil
}

func (s *Storage) AddToBlackList(ctx context.Context, ip storage.IP) error {
	query := `INSERT INTO blacklist (ip, mask)
VALUES($1, $2)`
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.DB.ExecContext(ctx, query, ip.IP, ip.Mask)
	if err != nil {
		return fmt.Errorf("query error %w", err)
	}

	s.l.Info("Added to black list", zap.String("ip", ip.IP))
	return nil
}

func (s *Storage) DeleteFromBlackList(ctx context.Context, ip storage.IP) error {
	query := `DELETE from blacklist WHERE ip = $1`
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.DB.ExecContext(ctx, query, ip.IP)
	if err != nil {
		return fmt.Errorf("query error %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected error %w", err)
	}
	if rowsAffected > 0 {
		s.l.Info("ip deleted from black list", zap.String("ip", ip.IP))
	} else {
		s.l.Info("ip does not exist in black list", zap.String("ip", ip.IP))
	}
	return nil
}

func (s *Storage) GetFromWhiteList(ip storage.IP) (bool, error) {
	var exists bool
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM whitelist WHERE ip = $1)", ip.IP)
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("error while getting ip from white list %w", err)
	} else if !exists {
		return false, nil
	}
	return true, nil
}

func (s *Storage) GetFromBlackList(ip storage.IP) (bool, error) {
	var exists bool
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.DB.QueryRow("SELECT EXISTS(SELECT * FROM blacklist WHERE ip = $1)", ip.IP)
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("error while getting ip from black list %w", err)
	} else if !exists {
		return false, nil
	}
	return true, nil
}

func (s *Storage) GetAllFromWhiteList(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var address storage.Address
	var list []string
	rows, err := s.DB.QueryContext(ctx, "SELECT * FROM whitelist")
	if err != nil {
		return nil, fmt.Errorf("query error %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(
			&address.ID,
			&address.IP,
			&address.Mask,
		); err != nil {
			return nil, fmt.Errorf("scan error %w", err)
		}
		list = append(list, address.IP)
	}
	return list, nil
}

func (s *Storage) GetAllFromBlackList(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var address storage.Address
	var list []string
	rows, err := s.DB.QueryContext(ctx, "SELECT * FROM blacklist")
	if err != nil {
		return nil, fmt.Errorf("query error %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(
			&address.ID,
			&address.IP,
			&address.Mask,
		); err != nil {
			return nil, fmt.Errorf("scan error %w", err)
		}
		list = append(list, address.IP)
	}
	return list, nil
}
