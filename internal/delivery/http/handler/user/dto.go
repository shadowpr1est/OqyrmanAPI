package user

type updateUserRequest struct {
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	FullName *string `json:"full_name"`
	// avatar_url убран — аватар обновляется только через POST /users/me/avatar
	// (multipart/form-data с загрузкой в MinIO).
	// Если оставить здесь, пользователь получает 200 OK но аватар не меняется —
	// вводящее в заблуждение поведение.
}

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	QRCode    string `json:"qr_code"`
	CreatedAt string `json:"created_at"`
}
type updateRoleRequest struct {
	Role      string  `json:"role"       binding:"required,oneof=Admin Staff User"`
	LibraryID *string `json:"library_id"` // обязателен для Staff, null для остальных
}
