package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorResponse — стандартный формат ошибки API.
// code  — машиночитаемая константа для фронтенда (switch/case).
// message — человекочитаемое сообщение для отображения пользователю.
// fields — только для validation_error: карта поле → сообщение.
type ErrorResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Error codes — все константы в одном месте.
const (
	CodeValidationError     = "validation_error"
	CodeInvalidCredentials  = "invalid_credentials"
	CodeEmailAlreadyTaken   = "email_already_taken"
	CodePhoneTaken          = "phone_taken"
	CodeEmailNotVerified    = "email_not_verified"
	CodeAlreadyVerified     = "already_verified"
	CodeInvalidCode         = "invalid_code"
	CodeRegistrationPending = "registration_pending"
	CodeNotFound            = "not_found"
	CodeForbidden           = "forbidden"
	CodeConflict            = "conflict"
	CodeInternalError       = "internal_error"
	CodeInvalidToken        = "invalid_token"
	CodeTokenExpired        = "token_expired"
	CodeInvalidGoogleToken  = "invalid_google_token"
	CodeTooManyRequests     = "too_many_requests"
)

func Err(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Code: code, Message: message})
}

func ValidationErr(c *gin.Context, err error) {
	resp := ErrorResponse{
		Code:    CodeValidationError,
		Message: "request validation failed",
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		fields := make(map[string]string, len(errs))
		for _, fe := range errs {
			fields[strings.ToLower(fe.Field())] = fieldMessage(fe)
		}
		resp.Fields = fields
		if len(errs) > 0 {
			resp.Message = fmt.Sprintf("%s: %s", strings.ToLower(errs[0].Field()), fieldMessage(errs[0]))
		}
	}
	c.JSON(http.StatusBadRequest, resp)
}

func BadRequest(c *gin.Context, code, message string) {
	Err(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code, message string) {
	Err(c, http.StatusUnauthorized, code, message)
}

func Forbidden(c *gin.Context) {
	Err(c, http.StatusForbidden, CodeForbidden, "access denied")
}

func NotFound(c *gin.Context, message string) {
	Err(c, http.StatusNotFound, CodeNotFound, message)
}

func Conflict(c *gin.Context, code, message string) {
	Err(c, http.StatusConflict, code, message)
}

func InternalError(c *gin.Context) {
	Err(c, http.StatusInternalServerError, CodeInternalError, "internal server error")
}
