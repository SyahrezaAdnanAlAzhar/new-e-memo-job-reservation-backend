package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) StoreRefreshToken(ctx context.Context, userID int, tokenID string, expiresAt time.Time) error {
	query := "INSERT INTO active_refresh_tokens (token_id, user_id, expires_at) VALUES ($1, $2, $3)"
	_, err := r.DB.ExecContext(ctx, query, tokenID, userID, expiresAt)
	return err
}

func (r *AuthRepository) ValidateAndDelRefreshToken(ctx context.Context, userID int, tokenID string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var exists bool
	queryCheck := "SELECT EXISTS(SELECT 1 FROM active_refresh_tokens WHERE token_id = $1 AND user_id = $2 AND expires_at > NOW())"
	err = tx.QueryRowContext(ctx, queryCheck, tokenID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("token not found, invalid, or expired")
	}

	queryDelete := "DELETE FROM active_refresh_tokens WHERE token_id = $1"
	_, err = tx.ExecContext(ctx, queryDelete, tokenID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AuthRepository) BlacklistToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	query := "INSERT INTO token_blacklist (token_id, expires_at) VALUES ($1, $2) ON CONFLICT (token_id) DO NOTHING"
	_, err := r.DB.ExecContext(ctx, query, tokenID, expiresAt)
	return err
}

func (r *AuthRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token_id = $1 AND expires_at > NOW())"
	err := r.DB.QueryRowContext(ctx, query, tokenID).Scan(&exists)
	return exists, err
}

func (r *AuthRepository) DeleteAllUserRefreshTokens(ctx context.Context, userID int) error {
	query := "DELETE FROM active_refresh_tokens WHERE user_id = $1"
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

func (r *AuthRepository) StoreWebSocketTicket(ctx context.Context, ticket string, userID int, expiresAt time.Time) error {
	query := "INSERT INTO websocket_tickets (ticket, user_id, expires_at) VALUES ($1, $2, $3)"
	
	// Use NULL for userID = 0 (public users)
	var userIDParam interface{}
	if userID == 0 {
		userIDParam = nil
		log.Printf("[DEBUG] Executing query: %s with ticket=%s, userID=NULL (public), expiresAt=%v", query, ticket, expiresAt)
	} else {
		userIDParam = userID
		log.Printf("[DEBUG] Executing query: %s with ticket=%s, userID=%d, expiresAt=%v", query, ticket, userID, expiresAt)
	}
	
	result, err := r.DB.ExecContext(ctx, query, ticket, userIDParam, expiresAt)
	if err != nil {
		log.Printf("[ERROR] Database error storing WebSocket ticket: %v", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] Successfully inserted WebSocket ticket, rows affected: %d", rowsAffected)
	return nil
}

func (r *AuthRepository) ValidateAndDelWebSocketTicket(ctx context.Context, ticket string) (userID int, err error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queryCheck := "SELECT user_id FROM websocket_tickets WHERE ticket = $1 AND expires_at > NOW()"
	var nullableUserID sql.NullInt64
	err = tx.QueryRowContext(ctx, queryCheck, ticket).Scan(&nullableUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("invalid or expired websocket ticket")
		}
		return 0, err
	}

	// Convert sql.NullInt64 to int, 0 if NULL (for public users)
	if nullableUserID.Valid {
		userID = int(nullableUserID.Int64)
	} else {
		userID = 0 // Public/anonymous user
	}

	queryDelete := "DELETE FROM websocket_tickets WHERE ticket = $1"
	_, err = tx.ExecContext(ctx, queryDelete, ticket)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	return userID, err
}

func (r *AuthRepository) SetEditMode(ctx context.Context, status bool) error {
	query := `
        INSERT INTO system_config (key, value) 
        VALUES ('edit_mode', $1) 
        ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`
	_, err := r.DB.ExecContext(ctx, query, fmt.Sprintf("%t", status))
	return err
}

func (r *AuthRepository) GetEditMode(ctx context.Context) (bool, error) {
	var value string
	query := "SELECT value FROM system_config WHERE key = 'edit_mode'"
	err := r.DB.QueryRowContext(ctx, query).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return value == "true", nil
}