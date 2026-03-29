package book_file

type bookFileResponse struct {
	ID      string `json:"id"`
	BookID  string `json:"book_id"`
	Format  string `json:"format"`
	FileURL string `json:"file_url"`
	IsAudio bool   `json:"is_audio"`
}
