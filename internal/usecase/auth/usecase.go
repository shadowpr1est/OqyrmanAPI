package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/ctxkeys"
	googlePkg "github.com/shadowpr1est/OqyrmanAPI/pkg/google"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/phone"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost       = 12
	maxLoginAttempts = 10
	lockoutDuration  = 15 * time.Minute
)

// EmailSender — минимальный интерфейс для отправки кода верификации.
// Позволяет подменять реализацию в тестах.
type EmailSender interface {
	Enabled() bool
	SendVerificationCode(to, code string) error
	SendPasswordResetCode(to, code string) error
}

type loginRecord struct {
	count         int
	lockedAt      *time.Time
	lastAttemptAt time.Time
}

type authUseCase struct {
	userRepo        repository.UserRepository
	tokenRepo       repository.TokenRepository
	verifRepo       repository.EmailVerificationCodeRepository
	resetRepo       repository.PasswordResetCodeRepository
	emailSender     EmailSender
	jwt             *jwt.Manager
	googleClientID  string
	otpSecret       []byte
	refreshTokenTTL time.Duration

	// защита от брутфорса: блокировка аккаунта после maxLoginAttempts неудачных попыток
	loginMu      sync.Mutex
	loginRecords map[string]*loginRecord
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	verifRepo repository.EmailVerificationCodeRepository,
	resetRepo repository.PasswordResetCodeRepository,
	emailSender EmailSender,
	jwt *jwt.Manager,
	googleClientID string,
	otpSecret string,
	refreshTokenTTLDays int,
) domainUseCase.AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		verifRepo:       verifRepo,
		resetRepo:       resetRepo,
		emailSender:     emailSender,
		jwt:             jwt,
		googleClientID:  googleClientID,
		otpSecret:       []byte(otpSecret),
		refreshTokenTTL: time.Duration(refreshTokenTTLDays) * 24 * time.Hour,
		loginRecords:    make(map[string]*loginRecord),
	}
}

func (u *authUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := validateEmail(user.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(user.PasswordHash); err != nil {
		return nil, err
	}
	normalized, err := phone.Normalize(user.Phone)
	if err != nil {
		return nil, err
	}
	user.Phone = normalized

	// Если email уже занят — перерегистрация разрешена только если аккаунт не верифицирован
	// И только если код верификации уже истёк (защита от DoS чужих аккаунтов)
	if existing, err := u.userRepo.GetByEmail(ctx, user.Email); err == nil {
		if existing.EmailVerified {
			return nil, entity.ErrEmailTaken
		}
		if err := u.checkReregistrationAllowed(ctx, existing.ID); err != nil {
			return nil, err
		}
		_ = u.userRepo.HardDelete(ctx, existing.ID)
	} else if !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("authUseCase.Register lookup email: %w", err)
	}

	// Если телефон занят — аналогичная логика
	if existing, err := u.userRepo.GetByPhone(ctx, user.Phone); err == nil {
		if existing.EmailVerified {
			return nil, entity.ErrPhoneTaken
		}
		if err := u.checkReregistrationAllowed(ctx, existing.ID); err != nil {
			return nil, err
		}
		_ = u.userRepo.HardDelete(ctx, existing.ID)
	} else if !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("authUseCase.Register lookup phone: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcryptCost)
	if err != nil {
		return nil, err
	}

	user.ID = uuid.New()
	user.PasswordHash = string(hash)
	user.Role = entity.RoleUser
	user.CreatedAt = time.Now()
	user.QRCode = generateQRToken()

	created, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	if err := u.sendCode(ctx, created); err != nil {
		return nil, err
	}

	return created, nil
}

func (u *authUseCase) SendVerificationCode(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrEmailNotFound
		}
		return fmt.Errorf("authUseCase.SendVerificationCode lookup: %w", err)
	}
	if user.EmailVerified {
		return entity.ErrAlreadyVerified
	}
	return u.sendCode(ctx, user)
}

func (u *authUseCase) VerifyEmail(ctx context.Context, email, code string) (*domainUseCase.TokenPair, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, entity.ErrEmailNotFound
		}
		return nil, fmt.Errorf("authUseCase.VerifyEmail lookup: %w", err)
	}
	if user.EmailVerified {
		return nil, entity.ErrAlreadyVerified
	}

	record, err := u.verifRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, entity.ErrCodeNotFound
	}
	if time.Now().After(record.ExpiresAt) {
		_ = u.verifRepo.DeleteByUserID(ctx, user.ID)
		return nil, entity.ErrCodeNotFound
	}
	if !u.verifyHMAC(code, record.Code) {
		return nil, entity.ErrCodeNotFound
	}

	if err := u.userRepo.SetEmailVerified(ctx, user.ID); err != nil {
		return nil, fmt.Errorf("authUseCase.VerifyEmail set verified: %w", err)
	}
	_ = u.verifRepo.DeleteByUserID(ctx, user.ID)

	user.EmailVerified = true

	return u.issueTokenPair(ctx, user)
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (*domainUseCase.TokenPair, error) {
	// Проверяем блокировку до обращения к БД — не раскрываем факт существования email
	if err := u.checkAndRecordFailedLogin(email, false); err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		_ = u.checkAndRecordFailedLogin(email, true)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		_ = u.checkAndRecordFailedLogin(email, true)
		return nil, errors.New("invalid credentials")
	}

	if !user.EmailVerified {
		// Успешная аутентификация — сбрасываем счётчик неудач
		u.resetLoginAttempts(email)
		return nil, entity.ErrEmailNotVerified
	}

	u.resetLoginAttempts(email)
	return u.issueTokenPair(ctx, user)
}

func (u *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	return u.tokenRepo.DeleteByRefreshToken(ctx, refreshToken)
}

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domainUseCase.TokenPair, error) {
	token, err := u.tokenRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if time.Now().After(token.ExpiresAt) {
		_ = u.tokenRepo.DeleteByRefreshToken(ctx, refreshToken)
		return nil, entity.ErrTokenExpired
	}

	user, err := u.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.tokenRepo.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	return u.issueTokenPair(ctx, user)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// checkReregistrationAllowed возвращает ошибку, если у неверифицированного пользователя
// есть активный (не истёкший) код верификации — в этом случае перерегистрация запрещена,
// чтобы нельзя было сбросить чужую регистрацию пока код ещё действует.
func (u *authUseCase) checkReregistrationAllowed(ctx context.Context, userID uuid.UUID) error {
	record, err := u.verifRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil
	}
	if time.Now().Before(record.ExpiresAt) {
		return entity.ErrRegistrationPending
	}
	return nil
}

// checkAndRecordFailedLogin проверяет блокировку (record=false) или фиксирует неудачную попытку (record=true).
// Возвращает entity.ErrAccountLocked если аккаунт заблокирован.
func (u *authUseCase) checkAndRecordFailedLogin(email string, record bool) error {
	u.loginMu.Lock()
	defer u.loginMu.Unlock()

	rec, exists := u.loginRecords[email]
	if !exists {
		if record {
			u.loginRecords[email] = &loginRecord{count: 1, lastAttemptAt: time.Now()}
		}
		return nil
	}

	// Блокировка истекла — сбрасываем
	if rec.lockedAt != nil && time.Since(*rec.lockedAt) >= lockoutDuration {
		delete(u.loginRecords, email)
		if record {
			u.loginRecords[email] = &loginRecord{count: 1, lastAttemptAt: time.Now()}
		}
		return nil
	}

	// Аккаунт заблокирован
	if rec.lockedAt != nil {
		return entity.ErrAccountLocked
	}

	// Запись устарела (давно не было попыток) — сбрасываем
	if time.Since(rec.lastAttemptAt) > lockoutDuration {
		delete(u.loginRecords, email)
		if record {
			u.loginRecords[email] = &loginRecord{count: 1, lastAttemptAt: time.Now()}
		}
		return nil
	}

	if record {
		rec.count++
		rec.lastAttemptAt = time.Now()
		if rec.count >= maxLoginAttempts {
			now := time.Now()
			rec.lockedAt = &now
		}
	}
	return nil
}

func (u *authUseCase) resetLoginAttempts(email string) {
	u.loginMu.Lock()
	delete(u.loginRecords, email)
	u.loginMu.Unlock()
}

// hmacCode вычисляет HMAC-SHA256 кода с серверным секретом.
// Используется для хранения OTP-кодов вместо bcrypt:
// 6-значные коды защищены TTL + rate limit + account lockout, bcrypt тут избыточен.
func (u *authUseCase) hmacCode(code string) string {
	mac := hmac.New(sha256.New, u.otpSecret)
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func (u *authUseCase) verifyHMAC(submitted, stored string) bool {
	expected := u.hmacCode(submitted)
	return hmac.Equal([]byte(expected), []byte(stored))
}

func (u *authUseCase) sendCode(ctx context.Context, user *entity.User) error {
	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("authUseCase.sendCode generate: %w", err)
	}

	record := &entity.EmailVerificationCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      u.hmacCode(code),
		ExpiresAt: time.Now().Add(3 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := u.verifRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("authUseCase.sendCode save: %w", err)
	}

	if u.emailSender != nil && u.emailSender.Enabled() {
		// Ошибка отправки письма не отменяет регистрацию: код уже сохранён в БД,
		// пользователь может запросить повтор через /auth/resend-code.
		if err := u.emailSender.SendVerificationCode(user.Email, code); err != nil {
			slog.WarnContext(ctx, "failed to send verification email",
				"err", err, "user_id", user.ID, "email", user.Email)
		}
	}

	return nil
}

func (u *authUseCase) issueTokenPair(ctx context.Context, user *entity.User) (*domainUseCase.TokenPair, error) {
	accessToken, err := u.jwt.GenerateAccessToken(user.ID, string(user.Role), user.LibraryID)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	userAgent, _ := ctx.Value(ctxkeys.UserAgentKey).(string)
	ip, _ := ctx.Value(ctxkeys.ClientIPKey).(string)
	token := &entity.Token{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(u.refreshTokenTTL),
		UserAgent:    userAgent,
		IP:           ip,
		CreatedAt:    time.Now(),
	}
	if err := u.tokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return &domainUseCase.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (u *authUseCase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Не раскрываем факт существования email
		return nil
	}

	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("authUseCase.ForgotPassword generate: %w", err)
	}

	record := &entity.PasswordResetCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      u.hmacCode(code),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := u.resetRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("authUseCase.ForgotPassword save: %w", err)
	}

	if u.emailSender != nil && u.emailSender.Enabled() {
		if err := u.emailSender.SendPasswordResetCode(user.Email, code); err != nil {
			slog.WarnContext(ctx, "failed to send password reset email",
				"err", err, "user_id", user.ID, "email", user.Email)
		}
	}

	return nil
}

func (u *authUseCase) ResendResetCode(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}
	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("authUseCase.ResendResetCode generate: %w", err)
	}
	record := &entity.PasswordResetCode{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      u.hmacCode(code),
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := u.resetRepo.Save(ctx, record); err != nil {
		return fmt.Errorf("authUseCase.ResendResetCode save: %w", err)
	}
	if u.emailSender != nil && u.emailSender.Enabled() {
		if err := u.emailSender.SendPasswordResetCode(user.Email, code); err != nil {
			slog.WarnContext(ctx, "failed to send resend reset email",
				"err", err, "user_id", user.ID, "email", user.Email)
		}
	}
	return nil
}

func (u *authUseCase) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if err := validatePassword(newPassword); err != nil {
		return err
	}

	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.New("invalid or expired code")
	}

	record, err := u.resetRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return entity.ErrResetCodeNotFound
	}
	if time.Now().After(record.ExpiresAt) {
		_ = u.resetRepo.DeleteByUserID(ctx, user.ID)
		return entity.ErrResetCodeNotFound
	}
	if !u.verifyHMAC(code, record.Code) {
		return entity.ErrResetCodeNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return err
	}

	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return fmt.Errorf("authUseCase.ResetPassword update: %w", err)
	}

	_ = u.resetRepo.DeleteByUserID(ctx, user.ID)
	_ = u.tokenRepo.DeleteAllByUserID(ctx, user.ID)

	return nil
}

func (u *authUseCase) LoginWithGoogle(ctx context.Context, idToken string) (*domainUseCase.TokenPair, error) {
	info, err := googlePkg.VerifyIDToken(idToken, u.googleClientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", entity.ErrValidation, err.Error())
	}

	// 1. Ищем по google_id
	user, err := u.userRepo.GetByGoogleID(ctx, info.Sub)
	if err == nil {
		return u.issueTokenPair(ctx, user)
	}
	if !errors.Is(err, entity.ErrNotFound) {
		return nil, err
	}

	// 2. Ищем по email — привязываем google_id к существующему аккаунту
	user, err = u.userRepo.GetByEmail(ctx, info.Email)
	if err == nil {
		if err := u.userRepo.SetGoogleID(ctx, user.ID, info.Sub); err != nil {
			return nil, err
		}
		user.GoogleID = &info.Sub
		if !user.EmailVerified {
			_ = u.userRepo.SetEmailVerified(ctx, user.ID)
			user.EmailVerified = true
		}
		return u.issueTokenPair(ctx, user)
	}
	if !errors.Is(err, entity.ErrNotFound) {
		return nil, err
	}

	// 3. Создаём нового пользователя через Google
	googlePhone := "g:" + info.Sub
	if len(googlePhone) > 20 {
		googlePhone = googlePhone[:20]
	}
	googleID := info.Sub
	newUser := &entity.User{
		ID:            uuid.New(),
		Email:         info.Email,
		Phone:         googlePhone,
		Name:          info.GivenName,
		Surname:       info.FamilyName,
		PasswordHash:  "",
		Role:          entity.RoleUser,
		GoogleID:      &googleID,
		AvatarURL:     info.Picture,
		EmailVerified: true,
		CreatedAt:     time.Now(),
	}
	newUser.QRCode = generateQRToken()

	created, err := u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("authUseCase.LoginWithGoogle create: %w", err)
	}

	return u.issueTokenPair(ctx, created)
}

func generateCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", n%1_000_000), nil
}

// generateQRToken создаёт криптографически случайный токен для QR-кода пользователя.
func generateQRToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("%w: invalid email format", entity.ErrValidation)
	}
	return nil
}

func validatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", entity.ErrValidation)
	}
	var hasUpper, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("%w: password must contain at least one uppercase letter", entity.ErrValidation)
	}
	if !hasDigit {
		return fmt.Errorf("%w: password must contain at least one digit", entity.ErrValidation)
	}
	return nil
}
