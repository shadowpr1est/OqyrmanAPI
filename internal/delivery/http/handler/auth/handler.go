package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.AuthUseCase
}

func NewHandler(uc domainUseCase.AuthUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Регистрация
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body registerRequest true "Данные для регистрации"
// @Success     201 {object} registerResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	user := &entity.User{
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: req.Password,
		Name:         req.Name,
		Surname:      req.Surname,
	}
	created, err := h.uc.Register(c.Request.Context(), user)
	if err != nil {
		if errors.Is(err, entity.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		if errors.Is(err, entity.ErrPhoneTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "phone already taken"})
			return
		}
		if errors.Is(err, entity.ErrRegistrationPending) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		slog.ErrorContext(c.Request.Context(), "register error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, registerResponse{
		Message: "verification code sent to your email",
		UserID:  created.ID.String(),
	})
}

// @Summary     Подтвердить email
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body verifyEmailRequest true "Email и 6-значный код"
// @Success     200 {object} tokenResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	pair, err := h.uc.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrCodeNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired code"})
		case errors.Is(err, entity.ErrAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{"error": "email already verified"})
		case errors.Is(err, entity.ErrEmailNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "verify email error", "err", err, "path", c.FullPath())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toAuthUserResponse(pair.User),
	})
}

// @Summary     Повторно отправить код
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body resendCodeRequest true "Email"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /auth/resend-code [post]
func (h *Handler) ResendCode(c *gin.Context) {
	var req resendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	if err := h.uc.SendVerificationCode(c.Request.Context(), req.Email); err != nil {
		switch {
		case errors.Is(err, entity.ErrAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{"error": "email already verified"})
		case errors.Is(err, entity.ErrEmailNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			slog.ErrorContext(c.Request.Context(), "resend code error", "err", err, "path", c.FullPath())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification code sent"})
}

// @Summary     Вход
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body loginRequest true "Email и пароль"
// @Success     200 {object} tokenResponse
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Router      /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, entity.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toAuthUserResponse(pair.User),
	})
}

// @Summary     Выход
// @Tags        auth
// @Security    BearerAuth
// @Accept      json
// @Param       input body logoutRequest true "Refresh токен"
// @Success     204
// @Failure     400 {object} map[string]string
// @Router      /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	if err := h.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Обновить токен
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body refreshRequest true "Refresh токен"
// @Success     200 {object} tokenResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	pair, err := h.uc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toAuthUserResponse(pair.User),
	})
}

// @Summary     Забыл пароль
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body forgotPasswordRequest true "Email"
// @Success     200 {object} map[string]string
// @Router      /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	// Всегда 200 — не раскрываем существование email
	_ = h.uc.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if this email exists, a reset code has been sent"})
}

// @Summary     Повторно отправить код сброса пароля
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body resendResetCodeRequest true "Email"
// @Success     200 {object} map[string]string
// @Router      /auth/resend-reset-code [post]
func (h *Handler) ResendResetCode(c *gin.Context) {
	var req resendResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	// Всегда 200 — не раскрываем существование email
	_ = h.uc.ResendResetCode(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if this email exists, a new reset code has been sent"})
}

// @Summary     Сброс пароля
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body resetPasswordRequest true "Email, код и новый пароль"
// @Success     200 {object} map[string]string
// @Failure     400 {object} map[string]string
// @Router      /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	if err := h.uc.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		if errors.Is(err, entity.ErrResetCodeNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired code"})
			return
		}
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		slog.ErrorContext(c.Request.Context(), "reset password error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// @Summary     Вход через Google
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body googleLoginRequest true "Google ID token"
// @Success     200 {object} tokenResponse
// @Failure     400 {object} map[string]string
// @Router      /auth/google [post]
func (h *Handler) LoginWithGoogle(c *gin.Context) {
	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	pair, err := h.uc.LoginWithGoogle(c.Request.Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Google token"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "google login error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toAuthUserResponse(pair.User),
	})
}

func toAuthUserResponse(u *entity.User) authUserResponse {
	if u == nil {
		return authUserResponse{}
	}
	return authUserResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Surname:   u.Surname,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Role:      string(u.Role),
	}
}
