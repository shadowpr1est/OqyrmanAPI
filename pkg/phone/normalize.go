package phone

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

// Normalize принимает номер в любом из форматов:
//
//	8 777 888 7788
//	+7 777 888 7788
//	777 888 7788
//	7 777 888 7788
//
// и возвращает его в виде +7XXXXXXXXXX.
// Возвращает entity.ErrValidation при неверном формате.
func Normalize(raw string) (string, error) {
	// Оставляем только цифры
	var b strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	digits := b.String()

	switch len(digits) {
	case 10:
		// 777 888 7788 → +77778887788
		digits = "7" + digits
	case 11:
		if digits[0] == '8' {
			// 8... → 7...
			digits = "7" + digits[1:]
		} else if digits[0] != '7' {
			return "", fmt.Errorf("%w: phone must start with 7 or 8", entity.ErrValidation)
		}
	default:
		return "", fmt.Errorf("%w: phone number must contain 10 or 11 digits", entity.ErrValidation)
	}

	return "+" + digits, nil
}
