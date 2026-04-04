package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

type conversationRepo struct {
	db *sqlx.DB
}

func NewConversationRepo(db *sqlx.DB) repository.ConversationRepository {
	return &conversationRepo{db: db}
}

func (r *conversationRepo) Create(ctx context.Context, conv *entity.Conversation) (*entity.Conversation, error) {
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO conversations (id, user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		conv.ID, conv.UserID, conv.Title, conv.CreatedAt, conv.UpdatedAt,
	).StructScan(conv)
	if err != nil {
		return nil, fmt.Errorf("conversationRepo.Create: %w", err)
	}
	return conv, nil
}

func (r *conversationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var c entity.Conversation
	err := r.db.GetContext(ctx, &c,
		`SELECT * FROM conversations WHERE id = $1`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrConversationNotFound
		}
		return nil, fmt.Errorf("conversationRepo.GetByID: %w", err)
	}
	return &c, nil
}

func (r *conversationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error) {
	var items []*entity.Conversation
	err := r.db.SelectContext(ctx, &items,
		`SELECT * FROM conversations WHERE user_id = $1 ORDER BY updated_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("conversationRepo.ListByUser: %w", err)
	}
	return items, nil
}

func (r *conversationRepo) UpdateTitle(ctx context.Context, id uuid.UUID, title string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET title = $1 WHERE id = $2`, title, id,
	)
	return err
}

func (r *conversationRepo) Touch(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversations SET updated_at = now() WHERE id = $1`, id,
	)
	return err
}

func (r *conversationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM conversations WHERE id = $1`, id,
	)
	return err
}

func (r *conversationRepo) SaveMessage(ctx context.Context, msg *entity.ChatMessage) (*entity.ChatMessage, error) {
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO chat_messages (id, conversation_id, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`,
		msg.ID, msg.ConversationID, msg.Role, msg.Content, msg.CreatedAt,
	).StructScan(msg)
	if err != nil {
		return nil, fmt.Errorf("conversationRepo.SaveMessage: %w", err)
	}
	return msg, nil
}

func (r *conversationRepo) ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*entity.ChatMessage, error) {
	// Берём последние N в обратном порядке, затем разворачиваем для хронологии
	var items []*entity.ChatMessage
	err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM (
			SELECT * FROM chat_messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) sub
		ORDER BY created_at ASC`,
		conversationID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("conversationRepo.ListMessages: %w", err)
	}
	return items, nil
}
