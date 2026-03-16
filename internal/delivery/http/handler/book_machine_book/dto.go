package book_machine_book

type createBookMachineBookRequest struct {
	MachineID       string `json:"machine_id"        binding:"required"`
	BookID          string `json:"book_id"           binding:"required"`
	TotalCopies     int    `json:"total_copies"      binding:"required"`
	AvailableCopies int    `json:"available_copies"  binding:"required"`
}

type updateBookMachineBookRequest struct {
	TotalCopies     *int `json:"total_copies"`
	AvailableCopies *int `json:"available_copies"`
}

type bookMachineBookResponse struct {
	ID              string `json:"id"`
	MachineID       string `json:"machine_id"`
	BookID          string `json:"book_id"`
	TotalCopies     int    `json:"total_copies"`
	AvailableCopies int    `json:"available_copies"`
}
