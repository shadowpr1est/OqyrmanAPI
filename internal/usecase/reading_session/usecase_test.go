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

// ─── Mock session repo ───────────────────────────────────────────────────────

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

// ─── Mock book repo ──────────────────────────────────────────────────────────

type mockBookRepo struct{ mock.Mock }

func (m *mockBookRepo) UpdateTotalPages(ctx context.Context, bookID uuid.UUID, totalPages int) error {
	return m.Called(ctx, bookID, totalPages).Error(0)
}

// Satisfy the rest of BookRepository interface with stubs.
func (m *mockBookRepo) Create(ctx context.Context, b *entity.Book) (*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) Update(ctx context.Context, b *entity.Book) (*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) UpdateCoverURL(ctx context.Context, id uuid.UUID, coverURL string) error {
	return nil
}
func (m *mockBookRepo) UpdateRating(ctx context.Context, bookID uuid.UUID) error { return nil }
func (m *mockBookRepo) ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockBookRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.BookView, error) {
	return nil, nil
}
func (m *mockBookRepo) ListView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListByAuthorView(ctx context.Context, authorID uuid.UUID) ([]*entity.BookView, error) {
	return nil, nil
}
func (m *mockBookRepo) ListByGenreView(ctx context.Context, genreID uuid.UUID) ([]*entity.BookView, error) {
	return nil, nil
}
func (m *mockBookRepo) SearchView(ctx context.Context, query string, limit, offset int) ([]*entity.BookView, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListPopularView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListSimilarView(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.BookView, error) {
	return nil, nil
}
func (m *mockBookRepo) ListRecommendedView(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.BookView, error) {
	return nil, nil
}

// ─── Upsert ───────────────────────────────────────────────────────────────────

func TestUpsert_GeneratesIDWhenNil(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	session := &entity.ReadingSession{
		UserID: uuid.New(),
		BookID: uuid.New(),
		Status: entity.StatusReading,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session, nil)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
}

func TestUpsert_KeepsExistingID(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	existingID := uuid.New()
	session := &entity.ReadingSession{
		ID:     existingID,
		UserID: uuid.New(),
		BookID: uuid.New(),
		Status: entity.StatusReading,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session, nil)

	assert.NoError(t, err)
	assert.Equal(t, existingID, session.ID)
}

func TestUpsert_SetsUpdatedAt(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	before := time.Now()
	session := &entity.ReadingSession{Status: entity.StatusReading}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session, nil)

	assert.NoError(t, err)
	assert.True(t, session.UpdatedAt.After(before) || session.UpdatedAt.Equal(before))
}

func TestUpsert_SetsFinishedAtWhenStatusFinished(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	before := time.Now()
	total := 100
	session := &entity.ReadingSession{
		UserID:      uuid.New(),
		BookID:      uuid.New(),
		CurrentPage: 100,
		Status:      entity.StatusFinished,
		FinishedAt:  nil,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)
	bookRepo.On("UpdateTotalPages", mock.Anything, session.BookID, total).Return(nil)

	_, err := uc.Upsert(context.Background(), session, &total)

	assert.NoError(t, err)
	assert.NotNil(t, session.FinishedAt)
	assert.True(t, session.FinishedAt.After(before) || session.FinishedAt.Equal(before))
}

func TestUpsert_DoesNotOverwriteFinishedAt(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	existing := time.Now().Add(-24 * time.Hour)
	total := 100
	session := &entity.ReadingSession{
		UserID:     uuid.New(),
		BookID:     uuid.New(),
		Status:     entity.StatusFinished,
		FinishedAt: &existing,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)
	bookRepo.On("UpdateTotalPages", mock.Anything, session.BookID, total).Return(nil)

	_, err := uc.Upsert(context.Background(), session, &total)

	assert.NoError(t, err)
	assert.Equal(t, &existing, session.FinishedAt)
}

func TestUpsert_NoFinishedAtForReadingStatus(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	session := &entity.ReadingSession{Status: entity.StatusReading}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	_, err := uc.Upsert(context.Background(), session, nil)

	assert.NoError(t, err)
	assert.Nil(t, session.FinishedAt)
}

func TestUpsert_ProtectsAgainstFalseFinished(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	session := &entity.ReadingSession{
		Status: entity.StatusFinished,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)

	// No total_pages → should downgrade to "reading"
	_, err := uc.Upsert(context.Background(), session, nil)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusReading, session.Status)
	assert.Nil(t, session.FinishedAt)
}

func TestUpsert_UpdatesTotalPagesOnBook(t *testing.T) {
	repo := new(mockSessionRepo)
	bookRepo := new(mockBookRepo)
	uc := readingSession.NewReadingSessionUseCase(repo, bookRepo)

	bookID := uuid.New()
	total := 250
	session := &entity.ReadingSession{
		BookID: bookID,
		Status: entity.StatusReading,
	}
	repo.On("Upsert", mock.Anything, session).Return(session, nil)
	bookRepo.On("UpdateTotalPages", mock.Anything, bookID, total).Return(nil)

	_, err := uc.Upsert(context.Background(), session, &total)

	assert.NoError(t, err)
	bookRepo.AssertCalled(t, "UpdateTotalPages", mock.Anything, bookID, total)
}
