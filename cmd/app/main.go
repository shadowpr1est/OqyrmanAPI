// @title           OqyrmanAPI
// @version         1.0
// @description     API для книжной платформы Oqyrman
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
//
// @tag.name          auth
// @tag.description   Регистрация, вход, выход и обновление токенов
//
// @tag.name          users
// @tag.description   Профиль пользователя и управление аккаунтами (admin)
//
// @tag.name          books
// @tag.description   Каталог книг — список, поиск, детали, управление (admin)
//
// @tag.name          authors
// @tag.description   Авторы — список, поиск, управление (admin)
//
// @tag.name          genres
// @tag.description   Жанры — список, управление (admin)
//
// @tag.name          libraries
// @tag.description   Библиотеки — список, поиск рядом, управление (admin)
//
// @tag.name          library-books
// @tag.description   Инвентарь библиотек — привязка книг и учёт копий
//
// @tag.name          book-files
// @tag.description   Файлы книг — PDF, EPUB, MP3
//
// @tag.name          events
// @tag.description   Мероприятия библиотек — список, детали, управление (admin)
//
// @tag.name          reservations
// @tag.description   Бронирование книг — пользователь, staff и admin операции
//
// @tag.name          reviews
// @tag.description   Отзывы и оценки книг
//
// @tag.name          reading-sessions
// @tag.description   Сессии чтения — прогресс и статус
//
// @tag.name          notes
// @tag.description   Заметки к книгам
//
// @tag.name          wishlist
// @tag.description   Список желаемых книг
//
// @tag.name          notifications
// @tag.description   Уведомления пользователя
//
// @tag.name          stats
// @tag.description   Статистика — платформа (admin), пользователь, библиотека (staff)
//
// @tag.name          ai
// @tag.description   AI-рекомендации и чат с книжным ассистентом
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/shadowpr1est/OqyrmanAPI/config"
	"github.com/shadowpr1est/OqyrmanAPI/docs"
	httpDelivery "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http"
	aiH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/ai"
	authH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/auth"
	authorH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/author"
	bookH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book"
	bookFileH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_file"
	eventH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/event"
	genreH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/genre"
	libraryH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library"
	libraryBookH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library_book"
	notificationH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/notification"
	notesH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reading_note"
	readingSessionH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reading_session"
	reservationH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reservation"
	reviewH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/review"
	statsH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/stats"
	userH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/user"
	wishlistH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/wishlist"
	"github.com/shadowpr1est/OqyrmanAPI/internal/repository/postgres"
	aiUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/ai"
	authUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/auth"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/email"
	authorUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/author"
	bookUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book"
	bookFileUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book_file"
	eventUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/event"
	genreUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/genre"
	libraryUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/library"
	libraryBookUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/library_book"
	notificationUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/notification"
	readingNoteUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reading_note"
	readingSessionUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reading_session"
	reservationUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reservation"
	reviewUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/review"
	statsUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/stats"
	userUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/user"
	wishlistUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/wishlist"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/hub"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm/anthropic"
	openaiLLM "github.com/shadowpr1est/OqyrmanAPI/pkg/llm/openai"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/logger"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/storage"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/worker"
)

func main() {
	// config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	logger.Init(cfg.App.Env)

	// Swagger host задаётся в рантайме из конфига — не хардкодится в аннотации.
	// Локально: SWAGGER_HOST=localhost:8080
	// На сервере: SWAGGER_HOST=<публичный_ip>:8080
	docs.SwaggerInfo.Host = cfg.App.SwaggerHost

	// db
	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("db error: %s", err)
	}

	// storage
	minioStorage, err := storage.NewMinioStorage(cfg)
	if err != nil {
		log.Fatalf("minio error: %s", err)
	}

	// jwt
	jwtManager, err := jwt.NewManager(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL)
	if err != nil {
		log.Fatalf("jwt error: %s", err)
	}

	// repositories
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokenRepo(db)
	verifRepo := postgres.NewEmailVerificationCodeRepo(db)
	resetRepo := postgres.NewPasswordResetCodeRepo(db)
	authorRepo := postgres.NewAuthorRepo(db)
	genreRepo := postgres.NewGenreRepo(db)
	bookRepo := postgres.NewBookRepo(db)
	bookFileRepo := postgres.NewBookFileRepo(db)
	sessionRepo := postgres.NewReadingSessionRepo(db)
	statsRepo := postgres.NewStatsRepo(db)
	wishlistRepo := postgres.NewWishlistRepo(db)
	noteRepo := postgres.NewReadingNoteRepo(db)
	libraryRepo := postgres.NewLibraryRepo(db)
	libraryBookRepo := postgres.NewLibraryBookRepo(db)
	reservationRepo := postgres.NewReservationRepo(db)
	reviewRepo := postgres.NewReviewRepo(db)
	notifRepo := postgres.NewNotificationRepo(db)
	eventRepo := postgres.NewEventRepo(db)
	convRepo := postgres.NewConversationRepo(db)

	// email sender (SMTP опционален — без него коды пишутся только в БД)
	emailSender := email.NewSender(
		cfg.Email.Host, cfg.Email.Port,
		cfg.Email.Username, cfg.Email.Password,
		cfg.Email.From,
	)

	// notification hub (SSE push)
	notifHub := hub.New()

	// usecases
	authUseCase := authUC.NewAuthUseCase(userRepo, tokenRepo, verifRepo, resetRepo, emailSender, jwtManager, cfg.Google.ClientID, cfg.JWT.RefreshTokenTTL)
	bookUseCase := bookUC.NewBookUseCase(bookRepo, minioStorage)
	userUseCase := userUC.NewUserUseCase(userRepo, tokenRepo, minioStorage)
	authorUseCase := authorUC.NewAuthorUseCase(authorRepo, minioStorage)
	genreUseCase := genreUC.NewGenreUseCase(genreRepo)
	bookFileUseCase := bookFileUC.NewBookFileUseCase(bookFileRepo, bookRepo, minioStorage)
	sessionUseCase := readingSessionUC.NewReadingSessionUseCase(sessionRepo)
	statsUseCase := statsUC.NewStatsUseCase(statsRepo)
	wishlistUseCase := wishlistUC.NewWishlistUseCase(wishlistRepo)
	noteUseCase := readingNoteUC.NewReadingNoteUseCase(noteRepo)
	libraryUseCase := libraryUC.NewLibraryUseCase(libraryRepo)
	libraryBookUseCase := libraryBookUC.NewLibraryBookUseCase(libraryBookRepo)
	reservUseCase := reservationUC.NewReservationUseCase(reservationRepo, notifRepo)
	reviewUseCase := reviewUC.NewReviewUseCase(reviewRepo, bookRepo)
	notifUseCase := notificationUC.NewNotificationUseCase(notifRepo, notifHub)
	eventUseCase := eventUC.NewEventUseCase(eventRepo, minioStorage)

	// AI
	var aiHandler *aiH.Handler
	var llmClient llm.LLMClient
	switch {
	case cfg.AI.OpenAIKey != "":
		llmClient = openaiLLM.NewClient(cfg.AI.OpenAIKey)
		slog.Info("AI: using OpenAI GPT")
	case cfg.AI.AnthropicKey != "":
		llmClient = anthropic.NewClient(cfg.AI.AnthropicKey)
		slog.Info("AI: using Anthropic Claude")
	default:
		slog.Info("AI: no API key set, AI endpoints disabled")
	}
	if llmClient != nil {
		aiUseCase := aiUC.NewAIUseCase(sessionRepo, wishlistRepo, bookRepo, convRepo, llmClient)
		aiHandler = aiH.NewHandler(aiUseCase)
	}

	// handlers
	authHandler := authH.NewHandler(authUseCase)
	userHandler := userH.NewHandler(userUseCase)
	authorHandler := authorH.NewHandler(authorUseCase)
	genreHandler := genreH.NewHandler(genreUseCase)
	bookHandler := bookH.NewHandler(bookUseCase)
	bookFileHandler := bookFileH.NewHandler(bookFileUseCase)
	sessionHandler := readingSessionH.NewHandler(sessionUseCase)
	statsHandler := statsH.NewHandler(statsUseCase)
	wishlistHandler := wishlistH.NewHandler(wishlistUseCase)
	notesHandler := notesH.NewHandler(noteUseCase)
	libraryHandler := libraryH.NewHandler(libraryUseCase)
	libraryBookHandler := libraryBookH.NewHandler(libraryBookUseCase)
	reservHandler := reservationH.NewHandler(reservUseCase)
	reviewHandler := reviewH.NewHandler(reviewUseCase)
	notifHandler := notificationH.NewHandler(notifUseCase, notifHub)
	eventHandler := eventH.NewHandler(eventUseCase)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	// background workers
	overdueCanceller := worker.NewOverdueCanceller(reservationRepo, 24*time.Hour)
	go overdueCanceller.Run(ctx)
	// router
	router := httpDelivery.NewRouter(
		ctx,
		db,
		minioStorage,
		authHandler,
		userHandler,
		authorHandler,
		genreHandler,
		bookHandler,
		bookFileHandler,
		sessionHandler,
		wishlistHandler,
		notesHandler,
		libraryHandler,
		libraryBookHandler,
		reservHandler,
		reviewHandler,
		jwtManager,
		statsHandler,
		notifHandler,
		aiHandler,
		eventHandler,
		cfg.App.Env,
		cfg.App.AllowedOrigins,
	)

	engine := router.Init()
	srv := &http.Server{
		Addr:         cfg.App.Host + ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "addr", cfg.App.Host+":"+cfg.App.Port, "swagger_host", cfg.App.SwaggerHost)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %s", err)
		}
	}()

	// ждём сигнала
	<-ctx.Done()
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %s", err)
	}
	slog.Info("server stopped")
}
