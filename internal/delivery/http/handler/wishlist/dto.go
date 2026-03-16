package wishlist

type addWishlistRequest struct {
	BookID string `json:"book_id" binding:"required"`
}

type wishlistResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	BookID  string `json:"book_id"`
	AddedAt string `json:"added_at"`
}
