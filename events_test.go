package tinybird

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/mock"
)

func newTestClient(mockClient *MockHttpClient) *ClientImpl {
	options := &ClientOptions{
		Protocol:   "https",
		Host:       "api.tinybird.co",
		ApiVersion: "v0",
		Token:      "test-token",
	}
	return NewClient(options, mockClient).(*ClientImpl)
}

func TestSendEvents_BasicSend(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"test","value":123}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=my_datasource"

	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		data,
		"application/x-ndjson",
		"",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "my_datasource", data, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_WithWaitOption(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"event":"click"}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=events&wait=true"

	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		data,
		"application/x-ndjson",
		"",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "events", data, &SendEventsOptions{
		Wait: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_WithJSONFormat(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"single","value":456}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=datasource&format=json"

	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		data,
		"application/json",
		"",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "datasource", data, &SendEventsOptions{
		Format: "json",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_WithGzipCompression(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"compressed","value":789}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=datasource"

	var capturedBody []byte
	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		mock.MatchedBy(func(body []byte) bool {
			capturedBody = body
			return true
		}),
		"application/x-ndjson",
		"gzip",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "datasource", data, &SendEventsOptions{
		Compress: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify body is gzip compressed by decompressing it
	reader, err := gzip.NewReader(bytes.NewReader(capturedBody))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !bytes.Equal(decompressed, data) {
		t.Errorf("decompressed body = %q, want %q", decompressed, data)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_WithZstdCompression(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"zstd-compressed","value":999}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=datasource"

	var capturedBody []byte
	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		mock.MatchedBy(func(body []byte) bool {
			capturedBody = body
			return true
		}),
		"application/x-ndjson",
		"zstd",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "datasource", data, &SendEventsOptions{
		Compress:            true,
		CompressionEncoding: "zstd",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify body is zstd compressed by decompressing it
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		t.Fatalf("failed to create zstd decoder: %v", err)
	}
	defer decoder.Close()

	decompressed, err := decoder.DecodeAll(capturedBody, nil)
	if err != nil {
		t.Fatalf("failed to decompress zstd: %v", err)
	}

	if !bytes.Equal(decompressed, data) {
		t.Errorf("decompressed body = %q, want %q", decompressed, data)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_UnsupportedCompression(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"test"}`)
	err := client.SendEvents(context.Background(), "datasource", data, &SendEventsOptions{
		Compress:            true,
		CompressionEncoding: "lz4",
	})

	if err == nil {
		t.Fatal("expected error for unsupported compression, got nil")
	}

	expectedErr := "unsupported compression encoding: lz4"
	if err.Error() != expectedErr {
		t.Errorf("error = %q, want %q", err.Error(), expectedErr)
	}
}

func TestSendEvents_WithAllOptions(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"full-options"}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=my_ds&wait=true&format=json"

	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		mock.MatchedBy(func(body []byte) bool {
			// Verify it's gzip compressed
			reader, err := gzip.NewReader(bytes.NewReader(body))
			if err != nil {
				return false
			}
			defer reader.Close()
			decompressed, err := io.ReadAll(reader)
			if err != nil {
				return false
			}
			return bytes.Equal(decompressed, data)
		}),
		"application/json",
		"gzip",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "my_ds", data, &SendEventsOptions{
		Wait:                true,
		Format:              "json",
		Compress:            true,
		CompressionEncoding: "gzip",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_URLEncodesDataSourceName(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{}`)
	expectedURL := "https://api.tinybird.co/v0/events?name=my+data+source"

	mockClient.On("PostRaw",
		mock.Anything,
		expectedURL,
		data,
		"application/x-ndjson",
		"",
		mock.AnythingOfType("*tinybird.WriteResponse"),
	).Return(nil)

	err := client.SendEvents(context.Background(), "my data source", data, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestSendEvents_ReturnsHTTPError(t *testing.T) {
	mockClient := NewMockHttpClient()
	client := newTestClient(mockClient)

	data := []byte(`{"name":"test"}`)
	expectedErr := errors.New("HTTP 500: Internal Server Error")

	mockClient.On("PostRaw",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(expectedErr)

	err := client.SendEvents(context.Background(), "datasource", data, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}

	mockClient.AssertExpectations(t)
}
