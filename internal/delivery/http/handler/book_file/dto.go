package book_file

type createBookFileRequest struct {
	BookID  string `json:"book_id"  binding:"required"`
	Format  string `json:"format"   binding:"required"` // pdf, epub, mp3
	FileURL string `json:"file_url" binding:"required"`
	IsAudio bool   `json:"is_audio"`
}

type bookFileResponse struct {
	ID      string `json:"id"`
	BookID  string `json:"book_id"`
	Format  string `json:"format"`
	FileURL string `json:"file_url"`
	IsAudio bool   `json:"is_audio"`
}
