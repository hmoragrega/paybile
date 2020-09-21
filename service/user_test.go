package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/hmoragrega/paybile/service"
	"github.com/hmoragrega/paybile/service/mocks"
)

func TestLogin(t *testing.T) {
	t.Parallel()
	var (
		ctx      = context.Background()
		dummyErr = fmt.Errorf("dummy error")
		login    = "foo"
		password = "bar"
		user     = service.User{ID: uuid.New(), HashedPassword: password}
	)
	tt := []struct {
		name     string
		login    string
		password string
		expect   func(*mocks.UserReader)
		wantUser service.User
		wantErr  error
	}{
		{
			name:     "reader error",
			login:    login,
			password: password,
			expect: func(r *mocks.UserReader) {
				r.On("GetByLogin", ctx, login).Return(service.User{}, dummyErr)
			},
			wantErr: dummyErr,
		}, {
			name:     "wrong password",
			login:    login,
			password: "wrong",
			expect: func(r *mocks.UserReader) {
				r.On("GetByLogin", ctx, login).Return(user, nil)
			},
			wantErr: service.ErrInvalidPassword,
		}, {
			name:     "login correct",
			login:    login,
			password: password,
			expect: func(r *mocks.UserReader) {
				r.On("GetByLogin", ctx, login).Return(user, nil)
			},
			wantUser: user,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ur := &mocks.UserReader{}
			tc.expect(ur)

			svc := service.UserService{
				Reader: ur,
			}

			got, err := svc.Login(ctx, tc.login, tc.password)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("unexpected error: got: %v, want %v", err, tc.wantErr)
			}
			if err == nil && !reflect.DeepEqual(got, tc.wantUser) {
				t.Fatalf("unexpected user: \n got:  %+v \n want: %+v", got, tc.wantUser)
			}
		})
	}
}
