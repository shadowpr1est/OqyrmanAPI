package user

type updateUserRequest struct {
	Name    *string `json:"name"`
	Surname *string `json:"surname"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	// avatar_url убран — аватар обновляется только через POST /users/me/avatar
	// (multipart/form-data с загрузкой в MinIO).
}

type userResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Name    string `json:"name"`
	Surname string `json:"surname"`

	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	QRCode    string `json:"qr_code"`
	CreatedAt string `json:"created_at"`
}

type adminUpdateUserRequest struct {
	Role      *string `json:"role"       binding:"omitempty,oneof=Admin Staff User"`
	LibraryID *string `json:"library_id"`
	Name      *string `json:"name"`
	Surname   *string `json:"surname"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
}

type createStaffRequest struct {
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=6"`
	LibraryID string `json:"library_id" binding:"required"`
	Name      string `json:"name"       binding:"required"`
	Surname   string `json:"surname"    binding:"required"`
	Phone     string `json:"phone"      binding:"required"`
}

type userViewResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Name        string  `json:"name"`
	Surname     string  `json:"surname"`
	AvatarURL   string  `json:"avatar_url,omitempty"`
	Role        string  `json:"role"`
	LibraryID   *string `json:"library_id,omitempty"`
	LibraryName string  `json:"library_name,omitempty"`
	CreatedAt   string  `json:"created_at"`
}
