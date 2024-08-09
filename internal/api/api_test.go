package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getUserType(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    string
		wantErr error
	}{
		{
			name:    "valid client token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJSb2xlIjoiY2xpZW50IiwiRW1haWwiOiIiLCJleHAiOjE3MjMzMTE0NzZ9.Q8i9MtidDRXtOkoNMZnVTFv2kQBug0a_mhlulksJG30",
			want:    "client",
			wantErr: nil,
		},
		{
			name:    "valid moderator token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJSb2xlIjoibW9kZXJhdG9yIiwiRW1haWwiOiIiLCJleHAiOjE3MjMzMDgxMjJ9.18gyTQprkH9y-tUZuaX_ODimkPWNplCpij-dnDUA3ok",
			want:    "moderator",
			wantErr: nil,
		},
		{
			name:    "invalid signing method",
			token:   "eywJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJSb2xlIjoibW9kZXJhdG9yIiwiRW1haWwiOiIiLCJleHAiOjE3MjMzMDgxMjJ9.18gyTQprkH9y-tUZuaX_ODimkPWNplCpij-dnDUA5ok",
			want:    "",
			wantErr: errFailedToCheckToken,
		},
		{
			name:    "invalid token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJSb2xlIjoibW9kZXJhdG9yIiwiRW1haWwiOiIiLCJleHAiOjE3MjMzMDgxMjJ9.18gyTQprkH9y-tUZuaX_ODimkPWNplCpij-dnDUA5ok",
			want:    "",
			wantErr: errFailedToCheckToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getUserType(tt.token)
			require.Equal(t, tt.want, got)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
