package phone_test

import (
	"testing"

	"github.com/shadowpr1est/OqyrmanAPI/pkg/phone"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"8 777 888 7788", "+77778887788", false},
		{"+7 777 888 7788", "+77778887788", false},
		{"777 888 7788", "+77778887788", false},
		{"7 777 888 7788", "+77778887788", false},
		{"87778887788", "+77778887788", false},
		{"+77778887788", "+77778887788", false},
		{"7778887788", "+77778887788", false},
		{"77778887788", "+77778887788", false},
		// invalid
		{"123", "", true},
		{"", "", true},
		{"9 777 888 7788", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := phone.Normalize(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
