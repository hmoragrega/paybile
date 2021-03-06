// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	service "github.com/hmoragrega/paybile/service"

	uuid "github.com/google/uuid"
)

// WalletGetter is an autogenerated mock type for the WalletGetter type
type WalletGetter struct {
	mock.Mock
}

// GetWallet provides a mock function with given fields: ctx, user, walletID
func (_m *WalletGetter) GetWallet(ctx context.Context, user service.User, walletID uuid.UUID) (service.Wallet, error) {
	ret := _m.Called(ctx, user, walletID)

	var r0 service.Wallet
	if rf, ok := ret.Get(0).(func(context.Context, service.User, uuid.UUID) service.Wallet); ok {
		r0 = rf(ctx, user, walletID)
	} else {
		r0 = ret.Get(0).(service.Wallet)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, service.User, uuid.UUID) error); ok {
		r1 = rf(ctx, user, walletID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
