package wishlist

import "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"

type addWishlistRequest struct {
	BookID string `json:"book_id" binding:"required"`
}

type wishlistResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	BookID  string `json:"book_id"`
	AddedAt string `json:"added_at"`
}

type wishlistViewResponse struct {
	ID      string          `json:"id"`
	UserID  string          `json:"user_id"`
	AddedAt string          `json:"added_at"`
	Book    common.BookRef  `json:"book"`
}
