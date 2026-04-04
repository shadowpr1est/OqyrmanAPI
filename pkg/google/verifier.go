package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var ErrInvalidToken = errors.New("invalid Google ID token")

// TokenInfo содержит данные пользователя из Google ID token.
type TokenInfo struct {
	Sub           string `json:"sub"`            // уникальный Google user ID
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"` // "true" / "false"
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Aud           string `json:"aud"` // должен совпадать с clientID
}

// VerifyIDToken проверяет Google ID token через tokeninfo endpoint.
// Если clientID пустой — проверка aud пропускается (удобно для тестов).
func VerifyIDToken(idToken, clientID string) (*TokenInfo, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)

	resp, err := http.Get(endpoint) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("google.VerifyIDToken: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidToken
	}

	var info TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("google.VerifyIDToken decode: %w", err)
	}

	if clientID != "" && info.Aud != clientID {
		return nil, ErrInvalidToken
	}
	if info.EmailVerified != "true" {
		return nil, errors.New("google: email not verified")
	}
	if info.Email == "" || info.Sub == "" {
		return nil, ErrInvalidToken
	}

	return &info, nil
}
