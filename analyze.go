package tinybird

import (
	"context"
	"fmt"
	"net/url"
)

func (c *ClientImpl) Analyze(ctx context.Context, input interface{}) (*AnalyzeResponse, error) {
	baseUrl := fmt.Sprintf("%s://%s/%s/analyze",
		c.options.Protocol,
		c.options.Host,
		c.options.ApiVersion,
	)

	var response AnalyzeResponse

	switch v := input.(type) {
	case []byte:
		// Local file upload via multipart form
		err := c.httpClient.PostMultipart(ctx, baseUrl, "file", "data", v, &response)
		if err != nil {
			return nil, err
		}
	case string:
		// Remote URL analysis
		reqUrl := baseUrl + "?url=" + url.QueryEscape(v)
		err := c.httpClient.PostRaw(ctx, reqUrl, nil, "", "", &response)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported input type: expected []byte or string, got %T", input)
	}

	return &response, nil
}
