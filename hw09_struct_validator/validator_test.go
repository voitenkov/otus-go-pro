package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version  string `validate:"nested"`
		AppToken Token  `validate:"nested"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	RequestUnknown struct {
		Code int `validate:"out:200,404,500"`
		Body string
	}

	RequestWrong struct {
		Code int
		Body string `validate:"max:10"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "1234567890",
				Name:   "Andrey",
				Age:    51,
				Email:  "test.com",
				Role:   "admin",
				Phones: []string{"12345678901", "12345678902"},
			},
			expectedErr: ErrLenValidator,
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "Andrey",
				Age:    51,
				Email:  "test@test.com",
				Role:   "admin",
				Phones: []string{"12345678901", "12345678902"},
			},
			expectedErr: ErrMaxValidator,
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "Andrey",
				Age:    50,
				Email:  "test",
				Role:   "admin",
				Phones: []string{"12345678901", "12345678902"},
			},
			expectedErr: ErrRegexpValidator,
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "Andrey",
				Age:    50,
				Email:  "test@test.ru",
				Role:   "admin",
				Phones: []string{"12345678901", "123"},
			},
			expectedErr: ErrLenValidator,
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "Andrey",
				Age:    50,
				Email:  "test@test.ru",
				Role:   "developer",
				Phones: []string{"12345678901", "12345678902"},
			},
			expectedErr: ErrInValidator,
		},
		{
			in: Response{
				Code: 999,
				Body: "test",
			},
			expectedErr: ErrInValidator,
		},
		{
			in: RequestUnknown{
				Code: 200,
				Body: "test",
			},
			expectedErr: ErrUnknownValidator,
		},
		{
			in: RequestWrong{
				Code: 200,
				Body: "test",
			},
			expectedErr: ErrValidatorMatching,
		},
		{
			in: App{
				Version:  "1.2.5",
				AppToken: Token{},
			},
			expectedErr: ErrNestedValidator,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			// t.Parallel()
			err := Validate(tt.in)
			require.Truef(t, errors.Is(err, tt.expectedErr), "actual error %q", err)
			_ = tt
		})
	}
}
