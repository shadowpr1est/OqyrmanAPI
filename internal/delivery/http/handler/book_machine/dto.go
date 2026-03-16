package book_machine

type createBookMachineRequest struct {
	Name    string  `json:"name"    binding:"required"`
	Address string  `json:"address" binding:"required"`
	Lat     float64 `json:"lat"     binding:"required"`
	Lng     float64 `json:"lng"     binding:"required"`
	Status  string  `json:"status"  binding:"required,oneof=active inactive"`
}

type updateBookMachineRequest struct {
	Name    *string  `json:"name"`
	Address *string  `json:"address"`
	Lat     *float64 `json:"lat"`
	Lng     *float64 `json:"lng"`
	Status  *string  `json:"status" binding:"omitempty,oneof=active inactive"`
}

type bookMachineResponse struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Status  string  `json:"status"`
}

type listBookMachineResponse struct {
	Items  []*bookMachineResponse `json:"items"`
	Total  int                    `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}
