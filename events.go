package tinybird

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/url"

	"github.com/klauspost/compress/zstd"
)

func (c *ClientImpl) SendEvents(ctx context.Context, datasourceName string, data []byte, options *SendEventsOptions) (*WriteResponse, error) {
	if options == nil {
		options = &SendEventsOptions{
			Wait:     false,
			Compress: false,
			Format:   "",
		}
	}

	// Build the URL for sending events
	reqUrl := fmt.Sprintf("%s://%s/%s/%s?name=%s",
		c.options.Protocol,
		c.options.Host,
		c.options.ApiVersion,
		"events",
		url.QueryEscape(datasourceName),
	)

	// Add query parameters based on options
	if options.Wait {
		reqUrl += "&wait=true"
	}
	if options.Format != "" {
		reqUrl += "&format=" + url.QueryEscape(options.Format)
	}

	// Prepare request body and content encoding
	body := data
	contentEncoding := ""

	if options.Compress {
		encoding := options.CompressionEncoding
		if encoding == "" {
			encoding = "gzip"
		}

		switch encoding {
		case "gzip":
			var buf bytes.Buffer
			gzWriter := gzip.NewWriter(&buf)
			if _, err := gzWriter.Write(data); err != nil {
				return nil, fmt.Errorf("failed to compress data: %w", err)
			}
			if err := gzWriter.Close(); err != nil {
				return nil, fmt.Errorf("failed to close gzip writer: %w", err)
			}
			body = buf.Bytes()
		case "zstd":
			encoder, err := zstd.NewWriter(nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
			}
			body = encoder.EncodeAll(data, nil)
			encoder.Close()
		default:
			return nil, fmt.Errorf("unsupported compression encoding: %s", encoding)
		}
		contentEncoding = encoding
	}

	// Set content type based on format
	contentType := "application/x-ndjson"
	if options.Format == "json" {
		contentType = "application/json"
	}

	var response WriteResponse

	err := c.httpClient.PostRaw(ctx, reqUrl, body, contentType, contentEncoding, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
