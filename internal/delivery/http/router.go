package http

import (
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	aiHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/ai"
	authHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/auth"
	authorHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/author"
	bookHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book"
	bookFileHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_file"
	bookMachineHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_machine"
	bookMachineBookHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/book_machine_book"
	genreHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/genre"
	libraryHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library"
	libraryBookHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/library_book"
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

type Router struct {
	auth            *authHandler.Handler
	user            *userHandler.Handler
	author          *authorHandler.Handler
	genre           *genreHandler.Handler
	book            *bookHandler.Handler
	bookFile        *bookFileHandler.Handler
	readingSession  *readingSessionHandler.Handler
	wishlist        *wishlistHandler.Handler
	notes           *notesHandler.Handler
	library         *libraryHandler.Handler
	libraryBook     *libraryBookHandler.Handler
	bookMachine     *bookMachineHandler.Handler
	bookMachineBook *bookMachineBookHandler.Handler
	reservation     *reservationHandler.Handler
	review          *reviewHandler.Handler
	jwt             *jwt.Manager
	stats           *statsHandler.Handler
	ai              *aiHandler.Handler
}

func NewRouter(
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
	bookMachine *bookMachineHandler.Handler,
	bookMachineBook *bookMachineBookHandler.Handler,
	reservation *reservationHandler.Handler,
	review *reviewHandler.Handler,
	jwt *jwt.Manager,
	stats *statsHandler.Handler,
	ai *aiHandler.Handler,
) *Router {
	return &Router{
		auth:            auth,
		user:            user,
		author:          author,
		genre:           genre,
		book:            book,
		bookFile:        bookFile,
		readingSession:  readingSession,
		wishlist:        wishlist,
		notes:           notes,
		library:         library,
		libraryBook:     libraryBook,
		bookMachine:     bookMachine,
		bookMachineBook: bookMachineBook,
		reservation:     reservation,
		review:          review,
		jwt:             jwt,
		stats:           stats,
		ai:              ai,
	}
}

func (r *Router) Init() *gin.Engine {
	engine := gin.Default()
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.auth.Register)
			auth.POST("/login", r.auth.Login)
			auth.POST("/refresh", r.auth.RefreshToken)
		}

		protected := api.Group("/")
		protected.Use(middleware.Auth(r.jwt))
		{
			protected.POST("/auth/logout", r.auth.Logout)
			protected.GET("/auth/me", r.auth.Me)

			// users
			protected.GET("/users/me", r.user.GetMe)
			protected.PUT("/users/me", r.user.Update)
			protected.DELETE("/users/me", r.user.Delete)
			protected.POST("/users/me/avatar", r.user.UploadAvatar)
			protected.GET("/users/me/qr", r.user.GetQR)

			// authors — /search ПЕРЕД /:id
			protected.GET("/authors/search", r.author.Search)
			protected.GET("/authors", r.author.List)
			protected.GET("/authors/:id", r.author.GetByID)

			// genres
			protected.GET("/genres", r.genre.List)
			protected.GET("/genres/:id", r.genre.GetByID)
			protected.GET("/genres/by-slug/:slug", r.genre.GetBySlug)

			// books — статичные маршруты ПЕРЕД /:id
			protected.GET("/books", r.book.List)
			protected.GET("/books/search", r.book.Search)
			protected.GET("/books/by-author/:author_id", r.book.ListByAuthor)
			protected.GET("/books/by-genre/:genre_id", r.book.ListByGenre)
			protected.GET("/books/:id", r.book.GetByID)
			protected.GET("/books/:id/availability", r.book.GetAvailability)

			// book files
			protected.GET("/book-files/:id", r.bookFile.GetByID)
			protected.GET("/book-files/by-book/:book_id", r.bookFile.ListByBook)

			// reading sessions
			protected.POST("/reading-sessions", r.readingSession.Upsert)
			protected.GET("/reading-sessions", r.readingSession.ListByUser)
			protected.GET("/reading-sessions/by-book/:book_id", r.readingSession.GetByBook)
			protected.DELETE("/reading-sessions/:id", r.readingSession.Delete)

			// wishlist
			protected.POST("/wishlist", r.wishlist.Add)
			protected.GET("/wishlist", r.wishlist.List)
			protected.GET("/wishlist/:book_id/exists", r.wishlist.Exists)
			protected.DELETE("/wishlist/:book_id", r.wishlist.Remove)

			// notes
			protected.POST("/notes", r.notes.Create)
			protected.GET("/notes/by-book/:book_id", r.notes.ListByBook)
			protected.GET("/notes/:id", r.notes.GetByID)
			protected.PUT("/notes/:id", r.notes.Update)
			protected.DELETE("/notes/:id", r.notes.Delete)

			// libraries
			protected.GET("/libraries", r.library.List)
			protected.GET("/libraries/nearby", r.library.ListNearby)
			protected.GET("/libraries/:id", r.library.GetByID)

			// library books
			protected.GET("/library-books/by-library/:library_id", r.libraryBook.ListByLibrary)
			protected.GET("/library-books/by-book/:book_id", r.libraryBook.ListByBook)
			protected.GET("/library-books/:id", r.libraryBook.GetByID)

			// book machines
			protected.GET("/book-machines", r.bookMachine.List)
			protected.GET("/book-machines/nearby", r.bookMachine.ListNearby)
			protected.GET("/book-machines/:id", r.bookMachine.GetByID)

			// book machine books
			protected.GET("/book-machine-books/by-machine/:machine_id", r.bookMachineBook.ListByMachine)
			protected.GET("/book-machine-books/by-book/:book_id", r.bookMachineBook.ListByBook)
			protected.GET("/book-machine-books/:id", r.bookMachineBook.GetByID)

			// reservations
			protected.POST("/reservations", r.reservation.Create)
			protected.GET("/reservations", r.reservation.ListByUser)
			protected.GET("/reservations/:id", r.reservation.GetByID)
			protected.PATCH("/reservations/:id/cancel", r.reservation.Cancel)

			// reviews
			protected.POST("/reviews", r.review.Create)
			protected.GET("/reviews/by-book/:book_id", r.review.ListByBook)
			protected.GET("/reviews/my", r.review.ListByUser)
			protected.GET("/reviews/:id", r.review.GetByID)
			protected.PUT("/reviews/:id", r.review.Update)
			protected.DELETE("/reviews/:id", r.review.Delete)

			// FIX: AI эндпоинты вынесены в отдельную группу с rate limit.
			// Каждый запрос к Claude API — платный. Без ограничений любой
			// авторизованный пользователь мог отправлять неограниченное число
			// запросов и опустошить Anthropic-баланс.
			// Лимит: 10 запросов в минуту с одного IP.
			if r.ai != nil {
				aiGroup := protected.Group("/ai")
				aiGroup.Use(middleware.RateLimit(10, time.Minute))
				{
					aiGroup.POST("/recommend", r.ai.Recommend)
					aiGroup.POST("/chat", r.ai.Chat)
				}
			}

			// admin
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				admin.GET("/stats", r.stats.GetStats)

				// users
				admin.GET("/users", r.user.ListAll)
				admin.PATCH("/users/:id/role", r.user.UpdateRole)
				admin.DELETE("/users/:id", r.user.AdminDelete)

				admin.POST("/authors", r.author.Create)
				admin.PUT("/authors/:id", r.author.Update)
				admin.DELETE("/authors/:id", r.author.Delete)

				admin.POST("/genres", r.genre.Create)
				admin.PUT("/genres/:id", r.genre.Update)
				admin.DELETE("/genres/:id", r.genre.Delete)

				admin.POST("/books", r.book.Create)
				admin.PUT("/books/:id", r.book.Update)
				admin.POST("/books/:id/cover", r.book.UploadCover)
				admin.DELETE("/books/:id", r.book.Delete)

				admin.POST("/book-files/upload", r.bookFile.Upload)
				admin.DELETE("/book-files/:id", r.bookFile.Delete)

				admin.POST("/libraries", r.library.Create)
				admin.PUT("/libraries/:id", r.library.Update)
				admin.DELETE("/libraries/:id", r.library.Delete)

				admin.POST("/library-books", r.libraryBook.Create)
				admin.PUT("/library-books/:id", r.libraryBook.Update)
				admin.DELETE("/library-books/:id", r.libraryBook.Delete)

				admin.POST("/book-machines", r.bookMachine.Create)
				admin.PUT("/book-machines/:id", r.bookMachine.Update)
				admin.DELETE("/book-machines/:id", r.bookMachine.Delete)

				admin.POST("/book-machine-books", r.bookMachineBook.Create)
				admin.PUT("/book-machine-books/:id", r.bookMachineBook.Update)
				admin.DELETE("/book-machine-books/:id", r.bookMachineBook.Delete)

				admin.PATCH("/reservations/:id/status", r.reservation.UpdateStatus)
				admin.PATCH("/reservations/:id/return", r.reservation.Return)
				admin.GET("/reservations", r.reservation.ListAll)
			}
		}
	}

	return engine
}
