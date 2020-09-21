//+build integration

package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/hmoragrega/paybile/service"
)

func TestGetByLogin(t *testing.T) {
	db := setupTestDB(t)
	tt := []struct {
		name     string
		login    string
		wantUser service.User
		wantErr  error
	}{
		{
			name:    "user not found",
			login:   "foo",
			wantErr: service.ErrUserNotFound,
		},
		{
			name:  "user found",
			login: "user_a",
			wantUser: service.User{
				ID:             userA.ID,
				Login:          "user_a",
				HashedPassword: "user_a_pass",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			setupFixtures(ctx, t, db)
			r := UserRepository{DB: db}

			got, err := r.GetByLogin(ctx, tc.login)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.wantUser) {
				t.Fatalf("unexpected user: \n got:  %+v \n want: %+v", got, tc.wantUser)
			}
		})
	}
}
