package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	aiHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/ai"
	authHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/auth"
	authorHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/author"
	bookHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book"
	bookFileHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_file"
	eventHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/event"
	genreHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/genre"
	libraryHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library"
	libraryBookHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library_book"
	notificationHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/notification"
	notesHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reading_note"
	readingSessionHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reading_session"
	reservationHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/reservation"
	reviewHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/review"
	statsHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/stats"
	userHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/user"
	wishlistHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/wishlist"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
)

type storagePinger interface {
	Ping(ctx context.Context) error
}

type Router struct {
	ctx            context.Context
	db             *sqlx.DB
	storage        storagePinger
	auth           *authHandler.Handler
	user           *userHandler.Handler
	author         *authorHandler.Handler
	genre          *genreHandler.Handler
	book           *bookHandler.Handler
	bookFile       *bookFileHandler.Handler
	readingSession *readingSessionHandler.Handler
	wishlist       *wishlistHandler.Handler
	notes          *notesHandler.Handler
	library        *libraryHandler.Handler
	libraryBook    *libraryBookHandler.Handler
	reservation    *reservationHandler.Handler
	review         *reviewHandler.Handler
	jwt            *jwt.Manager
	stats          *statsHandler.Handler
	notification   *notificationHandler.Handler
	ai             *aiHandler.Handler
	event          *eventHandler.Handler
	env            string
	allowedOrigins string
}

func NewRouter(
	ctx context.Context,
	db *sqlx.DB,
	storage storagePinger,
	auth *authHandler.Handler,
	user *userHandler.Handler,
	author *authorHandler.Handler,
	genre *genreHandler.Handler,
	book *bookHandler.Handler,
	bookFile *bookFileHandler.Handler,
	readingSession *readingSessionHandler.Handler,
	wishlist *wishlistHandler.Handler,
	notes *notesHandler.Handler,
	library *libraryHandler.Handler,
	libraryBook *libraryBookHandler.Handler,
	reservation *reservationHandler.Handler,
	review *reviewHandler.Handler,
	jwt *jwt.Manager,
	stats *statsHandler.Handler,
	notification *notificationHandler.Handler,
	ai *aiHandler.Handler,
	event *eventHandler.Handler,
	env string,
	allowedOrigins string,
) *Router {
	return &Router{
		ctx:            ctx,
		db:             db,
		storage:        storage,
		auth:           auth,
		user:           user,
		author:         author,
		genre:          genre,
		book:           book,
		bookFile:       bookFile,
		readingSession: readingSession,
		wishlist:       wishlist,
		notes:          notes,
		library:        library,
		libraryBook:    libraryBook,
		reservation:    reservation,
		review:         review,
		jwt:            jwt,
		stats:          stats,
		notification:   notification,
		ai:             ai,
		event:          event,
		env:            env,
		allowedOrigins: allowedOrigins,
	}
}

func (r *Router) Init() *gin.Engine {
	engine := gin.New()
	// Не доверяем прокси-заголовкам (X-Forwarded-For) по умолчанию.
	// В production за load balancer'ом — добавьте реальный IP прокси:
	//   engine.SetTrustedProxies([]string{"10.0.0.1"})
	_ = engine.SetTrustedProxies(nil)
	engine.Use(gin.Recovery())
	engine.Use(middleware.Metrics())
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.CORS(r.allowedOrigins))
	engine.Use(middleware.InjectRequestMeta())
	engine.MaxMultipartMemory = 20 << 20 // 20 MB
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	engine.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		if err := r.db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable"})
			return
		}
		if r.storage != nil {
			if err := r.storage.Ping(ctx); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "storage unavailable"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	if r.env != "production" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	api := engine.Group("/api/v1")
	rl := middleware.NewRateLimiter(r.ctx, time.Minute)
	{
		// ─── Публичные маршруты — без токена ───────────────────────────────
		public := api.Group("/")
		public.Use(middleware.RateLimitWithGroup(rl, "public", 100))
		{
			// auth — 20 req/min per IP
			authGroup := public.Group("/auth")
			authGroup.Use(middleware.RateLimitWithGroup(rl, "auth", 20))
			{
				authGroup.POST("/register", r.auth.Register)
				authGroup.POST("/verify-email", r.auth.VerifyEmail)
				authGroup.POST("/login", r.auth.Login)
				authGroup.POST("/refresh", r.auth.RefreshToken)
				authGroup.POST("/google", r.auth.LoginWithGoogle)
				// Чувствительные эндпоинты — 5 req/min per IP (брутфорс-защита)
				sensitive := authGroup.Group("/")
				sensitive.Use(middleware.RateLimitWithGroup(rl, "auth-sensitive", 5))
				{
					sensitive.POST("/resend-code", r.auth.ResendCode)
					sensitive.POST("/forgot-password", r.auth.ForgotPassword)
					sensitive.POST("/resend-reset-code", r.auth.ResendResetCode)
					sensitive.POST("/reset-password", r.auth.ResetPassword)
				}
			}

			// authors
			public.GET("/authors", r.author.List)
			public.GET("/authors/search", r.author.Search)
			public.GET("/authors/:id", r.author.GetByID)

			// genres
			public.GET("/genres", r.genre.List)
			public.GET("/genres/slug/:slug", r.genre.GetBySlug)
			public.GET("/genres/:id", r.genre.GetByID)

			// books — статичные маршруты должны быть ДО параметрических
			public.GET("/books", r.book.List)
			public.GET("/books/search", r.book.Search)
			public.GET("/books/popular", r.book.ListPopular)
			public.GET("/books/author/:author_id", r.book.ListByAuthor)
			public.GET("/books/genre/:genre_id", r.book.ListByGenre)
			public.GET("/books/:id", r.book.GetByID)
			public.GET("/books/:id/similar", r.book.ListSimilar)

			// libraries
			public.GET("/libraries", r.library.List)
			public.GET("/libraries/nearby", r.library.ListNearby)
			public.GET("/libraries/:id", r.library.GetByID)

			// events — публичные
			public.GET("/events", r.event.List)
			public.GET("/events/:id", r.event.GetByID)

			// reviews — читать без токена, писать только авторизованным
			public.GET("/reviews/book/:book_id", r.review.ListByBook)
		}

		// ─── Защищённые маршруты — нужен JWT токен ────────────────────────
		protected := api.Group("/")
		protected.Use(middleware.Auth(r.jwt))
		protected.Use(middleware.RateLimitWithGroup(rl, "protected", 60))
		{
			// auth
			protected.POST("/auth/logout", r.auth.Logout)

			// users
			protected.GET("/users/me", r.user.GetMe)
			protected.PUT("/users/me", r.user.Update)
			protected.DELETE("/users/me", r.user.Delete)
			protected.POST("/users/me/avatar", r.user.UploadAvatar)
			protected.GET("/users/me/qr", r.user.GetQR)
			protected.GET("/users/me/stats", r.stats.GetUserStats)
			protected.POST("/users/me/change-password", r.user.ChangePassword)
			protected.GET("/users/me/sessions", r.user.ListSessions)
			protected.DELETE("/users/me/sessions/:id", r.user.RevokeSession)
			protected.DELETE("/users/me/sessions", r.user.RevokeAllSessions)

			// book files — детальные данные только для авторизованных
			protected.GET("/book-files/:id", r.bookFile.GetByID)
			protected.GET("/book-files/book/:book_id", r.bookFile.ListByBook)

			// library books
			protected.GET("/library-books/library/:library_id", r.libraryBook.ListByLibrary)
			protected.GET("/library-books/book/:book_id", r.libraryBook.ListByBook)
			protected.GET("/library-books/:id", r.libraryBook.GetByID)

			// reading sessions
			protected.POST("/reading-sessions", r.readingSession.Upsert)
			protected.GET("/reading-sessions", r.readingSession.ListByUser)
			protected.GET("/reading-sessions/book/:book_id", r.readingSession.GetByBook)
			protected.DELETE("/reading-sessions/:id", r.readingSession.Delete)

			// wishlist
			protected.POST("/wishlist", r.wishlist.Add)
			protected.GET("/wishlist", r.wishlist.List)
			protected.GET("/wishlist/:book_id/exists", r.wishlist.Exists)
			protected.PATCH("/wishlist/:book_id/status", r.wishlist.UpdateStatus)
			protected.DELETE("/wishlist/:book_id", r.wishlist.Remove)

			// notes
			protected.POST("/notes", r.notes.Create)
			protected.GET("/notes/book/:book_id", r.notes.ListByBook)
			protected.GET("/notes/:id", r.notes.GetByID)
			protected.PUT("/notes/:id", r.notes.Update)
			protected.DELETE("/notes/:id", r.notes.Delete)

			// reservations
			protected.POST("/reservations", r.reservation.Create)
			protected.GET("/reservations", r.reservation.ListByUser)
			protected.GET("/reservations/:id", r.reservation.GetByID)
			protected.PATCH("/reservations/:id/cancel", r.reservation.Cancel)
			protected.PUT("/reservations/:id/extend", r.reservation.Extend)

			// notifications
			protected.GET("/notifications", r.notification.ListMy)
			protected.GET("/notifications/stream", r.notification.Stream)
			protected.PATCH("/notifications/:id/read", r.notification.MarkRead)
			protected.DELETE("/notifications/:id", r.notification.Delete)

			// reviews — писать и управлять только авторизованным
			protected.POST("/reviews", r.review.Create)
			protected.GET("/reviews/user", r.review.ListByUser)
			protected.GET("/reviews/:id", r.review.GetByID)
			protected.PUT("/reviews/:id", r.review.Update)
			protected.DELETE("/reviews/:id", r.review.Delete)

			// AI — rate limit 10 req/min на IP
			if r.ai != nil {
				aiGroup := protected.Group("/ai")
				aiGroup.Use(middleware.RateLimitWithGroup(rl, "ai", 10))
				{
					aiGroup.POST("/recommend", r.ai.Recommend)
					aiGroup.POST("/conversations", r.ai.CreateConversation)
					aiGroup.GET("/conversations", r.ai.ListConversations)
					aiGroup.GET("/conversations/:id", r.ai.GetConversation)
					aiGroup.POST("/conversations/:id/messages", r.ai.SendMessage)
					aiGroup.DELETE("/conversations/:id", r.ai.DeleteConversation)
				}
			}

			// ─── Admin маршруты — нужен JWT + роль Admin ──────────────────
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				admin.GET("/stats", r.stats.GetStats)

				// users
				admin.GET("/users", r.user.ListAll)
				admin.POST("/users/staff", r.user.CreateStaff)
				admin.PATCH("/users/:id", r.user.AdminUpdateUser)
				admin.DELETE("/users/:id", r.user.AdminDelete)

				// authors
				admin.POST("/authors", r.author.Create)
				admin.PUT("/authors/:id", r.author.Update)
				admin.DELETE("/authors/:id", r.author.Delete)

				// genres
				admin.POST("/genres", r.genre.Create)
				admin.PUT("/genres/:id", r.genre.Update)
				admin.DELETE("/genres/:id", r.genre.Delete)

				// books
				admin.POST("/books", r.book.Create)
				admin.PUT("/books/:id", r.book.Update)
				admin.POST("/books/:id/cover", r.book.UploadCover)
				admin.DELETE("/books/:id", r.book.Delete)

				// book files
				admin.POST("/book-files/upload", r.bookFile.Upload)
				admin.DELETE("/book-files/:id", r.bookFile.Delete)

				// libraries
				admin.POST("/libraries", r.library.Create)
				admin.PUT("/libraries/:id", r.library.Update)
				admin.DELETE("/libraries/:id", r.library.Delete)

				// library books
				admin.POST("/library-books", r.libraryBook.Create)
				admin.PUT("/library-books/:id", r.libraryBook.Update)
				admin.DELETE("/library-books/:id", r.libraryBook.Delete)

				// events
				admin.POST("/events", r.event.Create)
				admin.PUT("/events/:id", r.event.Update)
				admin.DELETE("/events/:id", r.event.Delete)

				// reservations
				admin.GET("/reservations", r.reservation.ListAll)
				admin.PATCH("/reservations/:id/status", r.reservation.UpdateStatus)
				admin.PATCH("/reservations/:id/return", r.reservation.AdminReturn)
			}

			// ─── Staff маршруты ───────────────────────────────────────────
			staff := protected.Group("/staff")
			staff.Use(middleware.StaffOnly())
			{
				staff.GET("/reservations", r.reservation.ListByLibrary)
				staff.PATCH("/reservations/:id/cancel", r.reservation.StaffCancel)
				staff.PATCH("/reservations/:id/return", r.reservation.StaffReturn)
				staff.PATCH("/reservations/:id/status", r.reservation.StaffUpdateStatus)
				staff.GET("/books/search", r.libraryBook.SearchInLibrary)
				staff.GET("/library/stats", r.stats.GetLibraryStats)
			}
		}
	}

	return engine
}
