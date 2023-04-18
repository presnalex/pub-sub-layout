package handler

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type StorageMock struct {
	mock.Mock
}

func (m *StorageMock) AnimalAdd(ctx context.Context, msg []byte) error {
	arguments := m.Called(ctx, msg)
	return arguments.Error(0)
}

func NewStorageMock() *StorageMock {
	return &StorageMock{}
}
