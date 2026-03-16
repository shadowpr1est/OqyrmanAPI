package user

type updateUserRequest struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
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
	Role string `json:"role" binding:"required,oneof=Admin User"`
}
