package auth

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Phone    string `json:"phone"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"     binding:"required"`
	Surname  string `json:"surname"  binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type authUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
}

type tokenResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	User         authUserResponse `json:"user"`
}

type verifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code"  binding:"required,len=6"`
}

type resendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type registerResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"        binding:"required,email"`
	Code        string `json:"code"         binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type googleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}
