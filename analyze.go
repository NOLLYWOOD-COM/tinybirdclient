package tinybird

import (
	"context"
	"fmt"
	"net/url"
)

func (c *ClientImpl) Analyze(ctx context.Context, input interface{}) (error, *AnalyzeResponse) {
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
			return err, nil
		}
	case string:
		// Remote URL analysis
		reqUrl := baseUrl + "?url=" + url.QueryEscape(v)
		err := c.httpClient.PostRaw(ctx, reqUrl, nil, "", "", &response)
		if err != nil {
			return err, nil
		}
	default:
		return fmt.Errorf("unsupported input type: expected []byte or string, got %T", input), nil
	}

	return nil, &response
}
