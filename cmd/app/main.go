// @title           OqyrmanAPI
// @version         1.0
// @description     API для книжной платформы Oqyrman
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @host localhost:8080
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
package main

import (
	"log"

	"github.com/shadowpr1est/OqyrmanAPI/config"
	_ "github.com/shadowpr1est/OqyrmanAPI/docs"
	httpDelivery "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http"
	aiH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/ai"
	authH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/auth"
	authorH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/author"
	bookH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book"
	bookFileH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_file"
	bookMachineH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_machine"
	bookMachineBookH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_machine_book"
	genreH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/genre"
	libraryH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library"
	libraryBookH "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library_book"
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
	authorUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/author"
	bookUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book"
	bookFileUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book_file"
	bookMachineUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book_machine"
	bookMachineBookUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/book_machine_book"
	genreUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/genre"
	libraryUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/library"
	libraryBookUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/library_book"
	readingNoteUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reading_note"
	readingSessionUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reading_session"
	reservationUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reservation"
	reviewUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/review"
	statsUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/stats"
	userUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/user"
	wishlistUC "github.com/shadowpr1est/OqyrmanAPI/internal/usecase/wishlist"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm/anthropic"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/storage"
)

func main() {
	// config
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

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
	jwtManager := jwt.NewManager(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL)

	// repositories
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewTokenRepo(db)
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
	machineRepo := postgres.NewBookMachineRepo(db)
	machineBookRepo := postgres.NewBookMachineBookRepo(db)
	reservationRepo := postgres.NewReservationRepo(db)
	reviewRepo := postgres.NewReviewRepo(db)

	// usecases
	authUseCase := authUC.NewAuthUseCase(userRepo, tokenRepo, jwtManager)
	bookUseCase := bookUC.NewBookUseCase(bookRepo, minioStorage)
	userUseCase := userUC.NewUserUseCase(userRepo, minioStorage)
	authorUseCase := authorUC.NewAuthorUseCase(authorRepo)
	genreUseCase := genreUC.NewGenreUseCase(genreRepo)
	bookFileUseCase := bookFileUC.NewBookFileUseCase(bookFileRepo, minioStorage)
	sessionUseCase := readingSessionUC.NewReadingSessionUseCase(sessionRepo)
	statsUseCase := statsUC.NewStatsUseCase(statsRepo)
	wishlistUseCase := wishlistUC.NewWishlistUseCase(wishlistRepo)
	noteUseCase := readingNoteUC.NewReadingNoteUseCase(noteRepo)
	libraryUseCase := libraryUC.NewLibraryUseCase(libraryRepo)
	libraryBookUseCase := libraryBookUC.NewLibraryBookUseCase(libraryBookRepo)
	machineUseCase := bookMachineUC.NewBookMachineUseCase(machineRepo)
	machineBookUseCase := bookMachineBookUC.NewBookMachineBookUseCase(machineBookRepo)

	// FIX: убраны libraryBookRepo и machineBookRepo — логика возврата/отмены копий
	// перенесена в reservationRepo (атомарные транзакции ReturnWithIncrement,
	// CancelWithIncrement). usecase больше не нуждается в этих зависимостях.
	reservUseCase := reservationUC.NewReservationUseCase(reservationRepo)

	reviewUseCase := reviewUC.NewReviewUseCase(reviewRepo)

	// AI
	var aiHandler *aiH.Handler
	if cfg.AI.AnthropicKey != "" {
		llmClient := anthropic.NewClient(cfg.AI.AnthropicKey)
		aiUseCase := aiUC.NewAIUseCase(sessionRepo, wishlistRepo, llmClient)
		aiHandler = aiH.NewHandler(aiUseCase)
	} else {
		log.Println("ANTHROPIC_API_KEY not set, AI endpoints disabled")
	}

	// handlers
	authHandler := authH.NewHandler(authUseCase)
	userHandler := userH.NewHandler(userUseCase)
	authorHandler := authorH.NewHandler(authorUseCase)
	genreHandler := genreH.NewHandler(genreUseCase)
	bookHandler := bookH.NewHandler(bookUseCase, libraryBookUseCase, machineBookUseCase)
	bookFileHandler := bookFileH.NewHandler(bookFileUseCase)
	sessionHandler := readingSessionH.NewHandler(sessionUseCase)
	statsHandler := statsH.NewHandler(statsUseCase)
	wishlistHandler := wishlistH.NewHandler(wishlistUseCase)
	notesHandler := notesH.NewHandler(noteUseCase)
	libraryHandler := libraryH.NewHandler(libraryUseCase)
	libraryBookHandler := libraryBookH.NewHandler(libraryBookUseCase)
	machineHandler := bookMachineH.NewHandler(machineUseCase)
	machineBookHandler := bookMachineBookH.NewHandler(machineBookUseCase)
	reservHandler := reservationH.NewHandler(reservUseCase)
	reviewHandler := reviewH.NewHandler(reviewUseCase)

	// router
	router := httpDelivery.NewRouter(
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
		machineHandler,
		machineBookHandler,
		reservHandler,
		reviewHandler,
		jwtManager,
		statsHandler,
		aiHandler,
	)

	engine := router.Init()

	log.Printf("server starting on %s:%s", cfg.App.Host, cfg.App.Port)
	if err := engine.Run(cfg.App.Host + ":" + cfg.App.Port); err != nil {
		log.Fatalf("server error: %s", err)
	}
}
