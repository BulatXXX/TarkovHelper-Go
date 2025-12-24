package repo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"tarkovhelper-api/internal/models"
)

type TrackedRepo struct {
	db *pgxpool.Pool
}

func NewTrackedRepo(db *pgxpool.Pool) *TrackedRepo { return &TrackedRepo{db: db} }

func (r *TrackedRepo) Get(ctx context.Context, userID string, mode models.Mode) ([]models.TrackedItem, error) {
	var raw []byte
	err := r.db.QueryRow(ctx, `
		SELECT items
		FROM tracked_items
		WHERE user_id = $1::uuid AND mode = $2
	`, userID, string(mode)).Scan(&raw)

	if err != nil {
		// если нет записи — вернём пусто, без ошибки
		// проще, чем тащить pgx.ErrNoRows наружу
		return []models.TrackedItem{}, nil
	}

	var items []models.TrackedItem
	if len(raw) == 0 {
		return []models.TrackedItem{}, nil
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TrackedRepo) Put(ctx context.Context, userID string, mode models.Mode, items []models.TrackedItem) ([]models.TrackedItem, error) {
	b, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO tracked_items (user_id, mode, items, updated_at)
		VALUES ($1::uuid, $2, $3::jsonb, $4)
		ON CONFLICT (user_id, mode)
		DO UPDATE SET items = EXCLUDED.items, updated_at = EXCLUDED.updated_at
	`, userID, string(mode), b, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return items, nil
}
