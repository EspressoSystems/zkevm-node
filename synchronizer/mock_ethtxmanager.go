// Code generated by mockery v2.16.0. DO NOT EDIT.

package synchronizer

import (
	context "context"

	pgx "github.com/jackc/pgx/v4"
	mock "github.com/stretchr/testify/mock"
)

// ethTxManagerMock is an autogenerated mock type for the ethTxManager type
type ethTxManagerMock struct {
	mock.Mock
}

// Reorg provides a mock function with given fields: ctx, fromBlockNumber, dbTx
func (_m *ethTxManagerMock) Reorg(ctx context.Context, fromBlockNumber uint64, dbTx pgx.Tx) error {
	ret := _m.Called(ctx, fromBlockNumber, dbTx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint64, pgx.Tx) error); ok {
		r0 = rf(ctx, fromBlockNumber, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTnewEthTxManagerMock interface {
	mock.TestingT
	Cleanup(func())
}

// newEthTxManagerMock creates a new instance of ethTxManagerMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func newEthTxManagerMock(t mockConstructorTestingTnewEthTxManagerMock) *ethTxManagerMock {
	mock := &ethTxManagerMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
