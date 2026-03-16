package http

import (
	"github.com/gin-gonic/gin"
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
	userHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/user"
	wishlistHandler "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/wishlist"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	ai              *aiHandler.Handler
	review          *reviewHandler.Handler
	jwt             *jwt.Manager
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
	ai *aiHandler.Handler,
	review *reviewHandler.Handler,
	jwt *jwt.Manager,
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
		ai:              ai,
		review:          review,
		jwt:             jwt,
	}
}

func (r *Router) Init() *gin.Engine {
	engine := gin.Default()

	// swagger
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api/v1")
	{
		// auth — публичные
		auth := api.Group("/auth")
		{
			auth.POST("/register", r.auth.Register)
			auth.POST("/login", r.auth.Login)
			auth.POST("/refresh", r.auth.RefreshToken)
		}

		// защищённые маршруты
		protected := api.Group("/")
		protected.Use(middleware.Auth(r.jwt))
		{
			// auth
			protected.POST("/auth/logout", r.auth.Logout)
			protected.GET("/auth/me", r.auth.Me)

			// user
			protected.GET("/users/me", r.user.GetMe)
			protected.GET("/users/me/qr", r.user.GetQR)
			protected.PUT("/users/me", r.user.Update)
			protected.DELETE("/users/me", r.user.Delete)

			// authors
			protected.GET("/authors", r.author.List)
			protected.GET("/authors/:id", r.author.GetByID)

			// genres
			protected.GET("/genres", r.genre.List)
			protected.GET("/genres/:id", r.genre.GetByID)
			protected.GET("/genres/slug/:slug", r.genre.GetBySlug)

			// books
			protected.GET("/books", r.book.List)
			protected.GET("/books/:id", r.book.GetByID)
			protected.GET("/books/:id/availability", r.book.GetAvailability)
			protected.GET("/books/search", r.book.Search)
			protected.GET("/books/author/:author_id", r.book.ListByAuthor)
			protected.GET("/books/genre/:genre_id", r.book.ListByGenre)

			// ai
			if r.ai != nil {
				protected.POST("/ai/recommend", r.ai.Recommend)
				protected.POST("/ai/chat", r.ai.Chat)
			}

			// book files
			protected.GET("/book-files/:id", r.bookFile.GetByID)
			protected.GET("/book-files/book/:book_id", r.bookFile.ListByBook)

			// reading sessions
			protected.POST("/reading-sessions", r.readingSession.Upsert)
			protected.GET("/reading-sessions", r.readingSession.ListByUser)
			protected.GET("/reading-sessions/book/:book_id", r.readingSession.GetByBook)
			protected.DELETE("/reading-sessions/:id", r.readingSession.Delete)

			// wishlist
			protected.POST("/wishlist", r.wishlist.Add)
			protected.GET("/wishlist", r.wishlist.List)
			protected.GET("/wishlist/:book_id/exists", r.wishlist.Exists)
			protected.DELETE("/wishlist/:book_id", r.wishlist.Remove)

			// notes
			protected.POST("/notes", r.notes.Create)
			protected.GET("/notes/:id", r.notes.GetByID)
			protected.GET("/notes/book/:book_id", r.notes.ListByBook)
			protected.PUT("/notes/:id", r.notes.Update)
			protected.DELETE("/notes/:id", r.notes.Delete)

			// libraries
			protected.GET("/libraries", r.library.List)
			protected.GET("/libraries/:id", r.library.GetByID)
			protected.GET("/libraries/nearby", r.library.ListNearby)

			// library books
			protected.GET("/library-books/:id", r.libraryBook.GetByID)
			protected.GET("/library-books/library/:library_id", r.libraryBook.ListByLibrary)
			protected.GET("/library-books/book/:book_id", r.libraryBook.ListByBook)

			// book machines
			protected.GET("/book-machines", r.bookMachine.List)
			protected.GET("/book-machines/:id", r.bookMachine.GetByID)
			protected.GET("/book-machines/nearby", r.bookMachine.ListNearby)

			// book machine books
			protected.GET("/book-machine-books/:id", r.bookMachineBook.GetByID)
			protected.GET("/book-machine-books/machine/:machine_id", r.bookMachineBook.ListByMachine)
			protected.GET("/book-machine-books/book/:book_id", r.bookMachineBook.ListByBook)

			// reservations
			protected.POST("/reservations", r.reservation.Create)
			protected.GET("/reservations/:id", r.reservation.GetByID)
			protected.GET("/reservations", r.reservation.ListByUser)
			protected.PATCH("/reservations/:id/cancel", r.reservation.Cancel)

			// reviews
			protected.POST("/reviews", r.review.Create)
			protected.GET("/reviews/:id", r.review.GetByID)
			protected.GET("/reviews/book/:book_id", r.review.ListByBook)
			protected.GET("/reviews/user", r.review.ListByUser)
			protected.PUT("/reviews/:id", r.review.Update)
			protected.DELETE("/reviews/:id", r.review.Delete)

			// admin маршруты
			admin := protected.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
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

				// book machines
				admin.POST("/book-machines", r.bookMachine.Create)
				admin.PUT("/book-machines/:id", r.bookMachine.Update)
				admin.DELETE("/book-machines/:id", r.bookMachine.Delete)

				// book machine books
				admin.POST("/book-machine-books", r.bookMachineBook.Create)
				admin.PUT("/book-machine-books/:id", r.bookMachineBook.Update)
				admin.DELETE("/book-machine-books/:id", r.bookMachineBook.Delete)

				// users
				admin.GET("/users", r.user.ListAll)
				admin.PATCH("/users/:id/role", r.user.UpdateRole)
				admin.DELETE("/users/:id", r.user.AdminDelete)

				// reservations
				admin.PATCH("/reservations/:id/return", r.reservation.Return)
			}
		}
	}

	return engine
}
