package tinybird

import (
	"context"
	"time"

	"github.com/NOLLYWOOD-COM/tinybird/internal/httpclient"
)

type Client interface {
	// Analyze analyzes the input data and provides a recommended schema and preview.
	//
	// ctx: The context for the request.
	//
	// input: The input data to be analyzed, a byte slice or remote file URL.
	//
	// Returns an error if the analysis fails, and an AnalyzeResponse containing the analysis results.
	Analyze(ctx context.Context, input interface{}) (*AnalyzeResponse, error)
	// CallEndpoint calls a Tinybird endpoint with the specified parameters.
	//
	// ctx: The context for the request.
	//
	// endpoint: The name of the Tinybird endpoint to call.
	//
	// params: A map of query parameters to include in the request.
	//
	// result: A pointer to a variable where the response will be unmarshaled.
	//
	// Returns an error if the request fails.
	CallEndpoint(ctx context.Context, endpoint string, params map[string]string) (*EndpointResponse, error)
	// SendEvents sends event data to the specified datasource.
	//
	// ctx: The context for the request.
	//
	// datasourceName: The name of the datasource to which events will be sent.
	//
	// data: The event data to be sent, typically in a byte slice format.
	//
	// options: Optional parameters for sending events, such as compression settings.
	SendEvents(ctx context.Context, datasourceName string, data []byte, options *SendEventsOptions) (*WriteResponse, error)
}

type SendEventsOptions struct {
	Wait                bool
	Compress            bool
	CompressionEncoding string // "gzip" or "zstd"
	Format              string // "json" for single JSON object, empty for NDJSON (default)
}

type ClientImpl struct {
	httpClient httpclient.Client
	options    *ClientOptions
}

type ClientOptions struct {
	Host       string
	Protocol   string
	ApiVersion string
	Token      string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

type Option func(*ClientOptions)

type WriteResponse struct {
	SuccessfulRows  int `json:"successful_rows"`  // Number of rows successfully written
	QuarantinedRows int `json:"quarantined_rows"` // Number of rows quarantined due to errors
}

type AnalyzeResponse struct {
	Analysis Analysis `json:"analysis"`
	Preview  Preview  `json:"preview"`
}

type Analysis struct {
	Columns []ColumnAnalysis `json:"columns"` // The columns attribute contains the guessed columns and for each one
	Schema  string           `json:"schema"`  // The recommended schema for the datasource
}

type ColumnAnalysis struct {
	Path            string `json:"path"`             // The JSON path to the column
	RecommendedType string `json:"recommended_type"` // The recommended data type for the column
	PresentPct      int    `json:"present_pct"`      // If the value is lower than 1 then there was nulls in the sample used for guessing
	Name            string `json:"name"`             // The recommended column name
}

type Preview struct {
	Meta  []FieldMeta                `json:"meta"`
	Data  []map[string][]interface{} `json:"data"`
	Rows  int                        `json:"rows"`
	Stats Statistics                 `json:"statistics"`
}

type FieldMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Statistics struct {
	Elasped   float64 `json:"elapsed"`
	RowsRead  int     `json:"rows_read"`
	BytesRead int     `json:"bytes_read"`
}

type EndpointResponse struct {
	Meta            []FieldMeta                `json:"meta"`
	Data            []map[string][]interface{} `json:"data"`
	Rows            int                        `json:"rows"`
	RowsBeforeLimit int                        `json:"rows_before_limit_at_least"`
	Stats           Statistics                 `json:"statistics"`
}
