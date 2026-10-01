package phase0

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AdminStore interface {
	CreateAdminIfAbsent(context.Context, string, string) error
}

func EnsureAdmin(ctx context.Context, store AdminStore, username, password string) error {
	if err := ValidateAdminCredentials(username, password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return store.CreateAdminIfAbsent(ctx, username, string(hash))
}

func ValidateAdminCredentials(username, password string) error {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return errors.New("admin username and password are required")
	}
	return nil
}
