package tinybird

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
)

func TestCallEndpoint_BasicCall(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	expectedURL := "https://api.tinybird.co/v0/pipes/my_endpoint"

	mockClient.On("Get",
		mock.Anything,
		expectedURL,
		mock.Anything,
		mock.AnythingOfType("*tinybird.EndpointResponse"),
	).Return(nil)

	_, err := client.CallEndpoint(context.Background(), "my_endpoint", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestCallEndpoint_WithParams(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	expectedURL := "https://api.tinybird.co/v0/pipes/analytics"
	params := map[string]string{
		"start_date": "2024-01-01",
		"end_date":   "2024-12-31",
		"limit":      "100",
	}

	mockClient.On("Get",
		mock.Anything,
		expectedURL,
		params,
		mock.AnythingOfType("*tinybird.EndpointResponse"),
	).Return(nil)

	_, err := client.CallEndpoint(context.Background(), "analytics", params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestCallEndpoint_ReturnsError(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	expectedErr := errors.New("network error")

	mockClient.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedErr)

	_, err := client.CallEndpoint(context.Background(), "failing_endpoint", nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}

	mockClient.AssertExpectations(t)
}

func TestCallEndpoint_EndpointWithSuffix(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	expectedURL := "https://api.tinybird.co/v0/pipes/my_pipe.json"

	mockClient.On("Get",
		mock.Anything,
		expectedURL,
		mock.Anything,
		mock.AnythingOfType("*tinybird.EndpointResponse"),
	).Return(nil)

	_, err := client.CallEndpoint(context.Background(), "my_pipe.json", nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestCallEndpoint_EmptyParams(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	expectedURL := "https://api.tinybird.co/v0/pipes/endpoint"
	emptyParams := map[string]string{}

	mockClient.On("Get",
		mock.Anything,
		expectedURL,
		emptyParams,
		mock.AnythingOfType("*tinybird.EndpointResponse"),
	).Return(nil)

	_, err := client.CallEndpoint(context.Background(), "endpoint", emptyParams)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestCallEndpoint_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	mockClient.On("Get",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(context.Canceled)

	_, err := client.CallEndpoint(ctx, "endpoint", nil)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}

	mockClient.AssertExpectations(t)
}
