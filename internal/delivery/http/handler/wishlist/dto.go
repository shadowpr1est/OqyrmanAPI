package wishlist

import "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"

type addWishlistRequest struct {
	BookID string `json:"book_id" binding:"required"`
	Status string `json:"status"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type wishlistResponse struct {
	ID      string `json:"id"`
	BookID  string `json:"book_id"`
	Status  string `json:"status"`
	AddedAt string `json:"added_at"`
}

type wishlistViewResponse struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	AddedAt string         `json:"added_at"`
	Book    common.BookRef `json:"book"`
}
