// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	service "github.com/hmoragrega/paybile/service"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// TransactionsReader is an autogenerated mock type for the TransactionsReader type
type TransactionsReader struct {
	mock.Mock
}

// ListTransactions provides a mock function with given fields: ctx, walletID, filters
func (_m *TransactionsReader) ListTransactions(ctx context.Context, walletID uuid.UUID, filters service.ListOptions) (service.TransactionList, error) {
	ret := _m.Called(ctx, walletID, filters)

	var r0 service.TransactionList
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, service.ListOptions) service.TransactionList); ok {
		r0 = rf(ctx, walletID, filters)
	} else {
		r0 = ret.Get(0).(service.TransactionList)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, service.ListOptions) error); ok {
		r1 = rf(ctx, walletID, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}