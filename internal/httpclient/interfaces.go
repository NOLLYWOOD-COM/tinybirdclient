package httpclient

import "context"

// Client is the internal HTTP client interface for making requests
type Client interface {
	Delete(ctx context.Context, url string, params map[string]string, result interface{}) error
	Get(ctx context.Context, url string, params map[string]string, result interface{}) error
	Patch(ctx context.Context, url string, body interface{}, result interface{}) error
	Post(ctx context.Context, url string, body interface{}, result interface{}) error
	PostMultipart(ctx context.Context, url string, fieldName string, fileName string, fileData []byte, result interface{}) error
	PostRaw(ctx context.Context, url string, body []byte, contentType string, contentEncoding string, result interface{}) error
	Put(ctx context.Context, url string, body interface{}, result interface{}) error
}
