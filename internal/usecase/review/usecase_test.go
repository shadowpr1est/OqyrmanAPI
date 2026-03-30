package review_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/usecase/review"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ─────────────────────────────────────────────────────────────────────

type mockReviewRepo struct{ mock.Mock }

func (m *mockReviewRepo) Create(ctx context.Context, r *entity.Review) (*entity.Review, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Review), args.Error(1)
}
func (m *mockReviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Review), args.Error(1)
}
func (m *mockReviewRepo) ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error) {
	args := m.Called(ctx, bookID, limit, offset)
	return args.Get(0).([]*entity.Review), args.Int(1), args.Error(2)
}
func (m *mockReviewRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.Review), args.Error(1)
}
func (m *mockReviewRepo) Update(ctx context.Context, r *entity.Review) (*entity.Review, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Review), args.Error(1)
}
func (m *mockReviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockReviewRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReviewView, error) {
	return nil, nil
}
func (m *mockReviewRepo) ListByBookView(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.ReviewView, int, error) {
	return nil, 0, nil
}
func (m *mockReviewRepo) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReviewView, error) {
	return nil, nil
}

// ─── mockBookRepo ─────────────────────────────────────────────────────────────

type mockBookRepo struct{ mock.Mock }

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
func (m *mockBookRepo) UpdateCoverURL(ctx context.Context, id uuid.UUID, url string) error { return nil }
func (m *mockBookRepo) UpdateRating(ctx context.Context, bookID uuid.UUID) error { return nil }
func (m *mockBookRepo) ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return nil, 0, nil
}
func (m *mockBookRepo) ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error) {
	return nil, nil
}
func (m *mockBookRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockBookRepo) UpdateTotalPages(ctx context.Context, bookID uuid.UUID, totalPages int) error {
	return nil
}
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

// ─── Create ───────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	repo := new(mockReviewRepo)
	bookRepo := new(mockBookRepo)
	uc := review.NewReviewUseCase(repo, bookRepo)

	input := &entity.Review{UserID: uuid.New(), BookID: uuid.New(), Rating: 5, Body: "Great!"}
	created := &entity.Review{ID: uuid.New(), UserID: input.UserID, BookID: input.BookID, Rating: 5, CreatedAt: time.Now()}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Review")).Return(created, nil)

	result, err := uc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, input.ID) // ID был проставлен usecase-ом
	repo.AssertExpectations(t)
}

func TestCreate_RatingTooLow(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	_, err := uc.Create(context.Background(), &entity.Review{Rating: 0})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 5")
	repo.AssertNotCalled(t, "Create")
}

func TestCreate_RatingTooHigh(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	_, err := uc.Create(context.Background(), &entity.Review{Rating: 6})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 5")
	repo.AssertNotCalled(t, "Create")
}

func TestCreate_BoundaryRatings(t *testing.T) {
	for _, rating := range []int{1, 5} {
		repo := new(mockReviewRepo)
		uc := review.NewReviewUseCase(repo, new(mockBookRepo))

		input := &entity.Review{Rating: rating}
		repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Review")).
			Return(&entity.Review{ID: uuid.New(), Rating: rating}, nil)

		_, err := uc.Create(context.Background(), input)
		assert.NoError(t, err, "rating %d should be valid", rating)
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Review")).
		Return(nil, errors.New("db error"))

	_, err := uc.Create(context.Background(), &entity.Review{Rating: 3})

	assert.Error(t, err)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	ownerID := uuid.New()
	input := &entity.Review{ID: uuid.New(), UserID: ownerID, Rating: 4}
	repo.On("Update", mock.Anything, input).Return(input, nil)

	result, err := uc.Update(context.Background(), input, ownerID)

	assert.NoError(t, err)
	assert.Equal(t, input, result)
}

func TestUpdate_Forbidden(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	input := &entity.Review{ID: uuid.New(), UserID: uuid.New(), Rating: 4}

	_, err := uc.Update(context.Background(), input, uuid.New()) // разные UserID

	assert.ErrorIs(t, err, entity.ErrForbidden)
	repo.AssertNotCalled(t, "Update")
}

func TestUpdate_InvalidRating(t *testing.T) {
	repo := new(mockReviewRepo)
	uc := review.NewReviewUseCase(repo, new(mockBookRepo))

	_, err := uc.Update(context.Background(), &entity.Review{Rating: 0}, uuid.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "between 1 and 5")
	repo.AssertNotCalled(t, "Update")
}
