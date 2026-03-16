package reservation

type createReservationRequest struct {
	LibraryBookID *string `json:"library_book_id"`
	MachineBookID *string `json:"machine_book_id"`
	SourceType    string  `json:"source_type" binding:"required,oneof=library machine"`
	DueDate       string  `json:"due_date"    binding:"required"` // "2006-01-02"
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending active completed cancelled"`
}

type reservationResponse struct {
	ID            string  `json:"id"`
	UserID        string  `json:"user_id"`
	LibraryBookID *string `json:"library_book_id"`
	MachineBookID *string `json:"machine_book_id"`
	SourceType    string  `json:"source_type"`
	Status        string  `json:"status"`
	ReservedAt    string  `json:"reserved_at"`
	DueDate       string  `json:"due_date"`
	ReturnedAt    *string `json:"returned_at"`
}
