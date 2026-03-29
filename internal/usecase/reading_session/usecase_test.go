package reading_session_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	readingSession "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reading_session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ─────────────────────────────────────────────────────────────────────

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Upsert(ctx context.Context, s *entity.ReadingSession) (*entity.ReadingSession, error) {
	args := m.Called(ctx, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ReadingSession), args.Error(1)
}
func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ReadingSession), args.Error(1)
}
func (m *mockSessionRepo) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error) {
	args := m.Called(ctx, userID, bookID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ReadingSession), args.Error(1)
}
func (m *mockSessionRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.ReadingSession), args.Error(1)
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSessionRepo) GetByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSessionView, error) {
	return nil, nil
}
func (m *mockSessionRepo) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSessionView, error) {
	return nil, nil
}

// ─── Upsert ───────────────────────────────────────────────────────────────────

func TestUpsert_GeneratesIDWhenNil(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	session := &entity.ReadingSession{
		UserID: uuid.New(),
		BookID: uuid.New(),
		Status: entity.StatusReading,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID) // ID был сгенерирован
}

func TestUpsert_KeepsExistingID(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	existingID := uuid.New()
	session := &entity.ReadingSession{
		ID:     existingID,
		UserID: uuid.New(),
		BookID: uuid.New(),
		Status: entity.StatusReading,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.Equal(t, existingID, session.ID) // ID не изменился
}

func TestUpsert_SetsUpdatedAt(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	before := time.Now()
	session := &entity.ReadingSession{Status: entity.StatusReading}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.True(t, session.UpdatedAt.After(before) || session.UpdatedAt.Equal(before))
}

func TestUpsert_SetsFinishedAtWhenStatusFinished(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	before := time.Now()
	session := &entity.ReadingSession{
		Status:     entity.StatusFinished,
		FinishedAt: nil, // не выставлен
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.NotNil(t, session.FinishedAt)
	assert.True(t, session.FinishedAt.After(before) || session.FinishedAt.Equal(before))
}

func TestUpsert_DoesNotOverwriteFinishedAt(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	existing := time.Now().Add(-24 * time.Hour)
	session := &entity.ReadingSession{
		Status:     entity.StatusFinished,
		FinishedAt: &existing, // уже выставлен
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.Equal(t, &existing, session.FinishedAt) // не изменился
}

func TestUpsert_NoFinishedAtForReadingStatus(t *testing.T) {
	repo := new(mockSessionRepo)
	uc := readingSession.NewReadingSessionUseCase(repo)

	session := &entity.ReadingSession{Status: entity.StatusReading}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session)

	assert.NoError(t, err)
	assert.Nil(t, session.FinishedAt)
}
