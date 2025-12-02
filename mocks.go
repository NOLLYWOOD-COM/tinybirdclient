package tinybird

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockHttpClient struct {
	mock.Mock
}

func NewMockHttpClient() *MockHttpClient {
	return &MockHttpClient{}
}

func (m *MockHttpClient) Delete(ctx context.Context, url string, params map[string]string, result interface{}) error {
	args := m.Called(ctx, url, params, result)
	return args.Error(0)
}

func (m *MockHttpClient) Get(ctx context.Context, url string, params map[string]string, result interface{}) error {
	args := m.Called(ctx, url, params, result)
	return args.Error(0)
}

func (m *MockHttpClient) Patch(ctx context.Context, url string, body interface{}, result interface{}) error {
	args := m.Called(ctx, url, body, result)
	return args.Error(0)
}

func (m *MockHttpClient) Post(ctx context.Context, url string, body interface{}, result interface{}) error {
	args := m.Called(ctx, url, body, result)
	return args.Error(0)
}

func (m *MockHttpClient) PostMultipart(ctx context.Context, url string, fieldName string, fileName string, fileData []byte, result interface{}) error {
	args := m.Called(ctx, url, fieldName, fileName, fileData, result)
	return args.Error(0)
}

func (m *MockHttpClient) PostRaw(ctx context.Context, url string, body []byte, contentType string, contentEncoding string, result interface{}) error {
	args := m.Called(ctx, url, body, contentType, contentEncoding, result)
	return args.Error(0)
}

func (m *MockHttpClient) Put(ctx context.Context, url string, body interface{}, result interface{}) error {
	args := m.Called(ctx, url, body, result)
	return args.Error(0)
}
