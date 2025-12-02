package tinybird

import (
	"context"
	"fmt"
)

func (c *ClientImpl) CallEndpoint(ctx context.Context, endpointName string, params map[string]string, result interface{}) error {
	reqUrl := fmt.Sprintf("%s://%s/%s/pipes/%s",
		c.options.Protocol,
		c.options.Host,
		c.options.ApiVersion,
		endpointName,
	)

	var response EndpointResponse

	err := c.httpClient.Get(ctx, reqUrl, params, &response)
	if err != nil {
		return err
	}

	return nil
}
