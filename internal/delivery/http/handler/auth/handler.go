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
// @Failure     400 {object} common.ErrorResponse
// @Failure     409 {object} common.ErrorResponse
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
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
		switch {
		case errors.Is(err, entity.ErrEmailTaken):
			common.Conflict(c, common.CodeEmailAlreadyTaken, "this email is already registered")
		case errors.Is(err, entity.ErrPhoneTaken):
			common.Conflict(c, common.CodePhoneTaken, "this phone number is already registered")
		case errors.Is(err, entity.ErrRegistrationPending):
			common.Conflict(c, common.CodeRegistrationPending, "verification code is still active, please check your email or wait 3 minutes")
		case errors.Is(err, entity.ErrValidation):
			common.BadRequest(c, common.CodeValidationError, err.Error())
		default:
			slog.ErrorContext(c.Request.Context(), "register error", "err", err, "path", c.FullPath())
			common.InternalError(c)
		}
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
// @Failure     400 {object} common.ErrorResponse
// @Failure     409 {object} common.ErrorResponse
// @Router      /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	pair, err := h.uc.VerifyEmail(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrCodeNotFound):
			common.BadRequest(c, common.CodeInvalidCode, "invalid or expired verification code")
		case errors.Is(err, entity.ErrAlreadyVerified):
			common.Conflict(c, common.CodeAlreadyVerified, "email is already verified")
		case errors.Is(err, entity.ErrEmailNotFound):
			common.NotFound(c, "user not found")
		default:
			slog.ErrorContext(c.Request.Context(), "verify email error", "err", err, "path", c.FullPath())
			common.InternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         toAuthUserResponse(pair.User),
	})
}

// @Summary     Повторно отправить код верификации
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body resendCodeRequest true "Email"
// @Success     200 {object} map[string]string
// @Failure     400 {object} common.ErrorResponse
// @Failure     409 {object} common.ErrorResponse
// @Router      /auth/resend-code [post]
func (h *Handler) ResendCode(c *gin.Context) {
	var req resendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if err := h.uc.SendVerificationCode(c.Request.Context(), req.Email); err != nil {
		switch {
		case errors.Is(err, entity.ErrAlreadyVerified):
			common.Conflict(c, common.CodeAlreadyVerified, "email is already verified")
		case errors.Is(err, entity.ErrEmailNotFound):
			common.NotFound(c, "user not found")
		default:
			slog.ErrorContext(c.Request.Context(), "resend code error", "err", err, "path", c.FullPath())
			common.InternalError(c)
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
// @Failure     401 {object} common.ErrorResponse
// @Failure     403 {object} common.ErrorResponse
// @Router      /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrEmailNotVerified):
			common.Err(c, http.StatusForbidden, common.CodeEmailNotVerified, "please verify your email before logging in")
		case errors.Is(err, entity.ErrAccountLocked):
			c.Header("Retry-After", "900")
			common.Err(c, http.StatusTooManyRequests, common.CodeTooManyRequests, "account temporarily locked, try again in 15 minutes")
		default:
			common.Unauthorized(c, common.CodeInvalidCredentials, "invalid email or password")
		}
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
// @Failure     400 {object} common.ErrorResponse
// @Router      /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if err := h.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
// @Failure     401 {object} common.ErrorResponse
// @Router      /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	pair, err := h.uc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, entity.ErrTokenExpired) {
			common.Unauthorized(c, common.CodeTokenExpired, "refresh token has expired")
			return
		}
		common.Unauthorized(c, common.CodeInvalidToken, "invalid refresh token")
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
		common.ValidationErr(c, err)
		return
	}

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
		common.ValidationErr(c, err)
		return
	}

	_ = h.uc.ResendResetCode(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if this email exists, a new reset code has been sent"})
}

// @Summary     Сброс пароля
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body resetPasswordRequest true "Email, код и новый пароль"
// @Success     200 {object} map[string]string
// @Failure     400 {object} common.ErrorResponse
// @Router      /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if err := h.uc.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, entity.ErrResetCodeNotFound):
			common.BadRequest(c, common.CodeInvalidCode, "invalid or expired reset code")
		case errors.Is(err, entity.ErrValidation):
			common.BadRequest(c, common.CodeValidationError, err.Error())
		default:
			slog.ErrorContext(c.Request.Context(), "reset password error", "err", err, "path", c.FullPath())
			common.InternalError(c)
		}
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
// @Failure     400 {object} common.ErrorResponse
// @Router      /auth/google [post]
func (h *Handler) LoginWithGoogle(c *gin.Context) {
	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	pair, err := h.uc.LoginWithGoogle(c.Request.Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, entity.ErrValidation) {
			common.BadRequest(c, common.CodeInvalidGoogleToken, "invalid Google token")
			return
		}
		slog.ErrorContext(c.Request.Context(), "google login error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
