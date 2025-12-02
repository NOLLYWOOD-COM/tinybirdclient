# Tinybird Go Client

A Go client library for the [Tinybird](https://www.tinybird.co/) analytics platform.

## Installation

```bash
go get github.com/NOLLYWOOD-COM/tinybird
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    tinybird "github.com/NOLLYWOOD-COM/tinybird"
)

func main() {
    // Create client with default options (reads token from TINYBIRD_TOKEN env var)
    client := tinybird.NewClient(tinybird.DefaultClientOptions(), nil)

    // Send events to a datasource
    data := []byte(`{"event":"page_view","user_id":"123"}`)
    err := client.SendEvents(context.Background(), "events_ds", data, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

### Using Functional Options

```go
options := tinybird.NewClientOptions(
    tinybird.Token("your-api-token"),
    tinybird.Host("api.tinybird.co"),
    tinybird.Timeout(30 * time.Second),
    tinybird.MaxRetries(5),
    tinybird.RetryDelay(3 * time.Second),
)

client := tinybird.NewClient(options, nil)
```

### Using Default Options

```go
// Uses environment variable TINYBIRD_TOKEN for authentication
client := tinybird.NewClient(tinybird.DefaultClientOptions(), nil)
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `Token(string)` | `$TINYBIRD_TOKEN` | API authentication token |
| `Host(string)` | `api.tinybird.co` | Tinybird API host |
| `Protocol(string)` | `https` | HTTP protocol |
| `ApiVersion(string)` | `v0` | API version |
| `Timeout(time.Duration)` | `15s` | Request timeout |
| `MaxRetries(int)` | `3` | Maximum retry attempts for failed requests |
| `RetryDelay(time.Duration)` | `2s` | Delay between retries |

## API Reference

### SendEvents

Send event data to a Tinybird datasource.

```go
func (c *Client) SendEvents(
    ctx context.Context,
    datasourceName string,
    data []byte,
    options *SendEventsOptions,
) error
```

#### SendEventsOptions

| Field | Type | Description |
|-------|------|-------------|
| `Wait` | `bool` | Wait for write acknowledgment before returning |
| `Compress` | `bool` | Enable compression for the request body |
| `CompressionEncoding` | `string` | Compression algorithm: `"gzip"` (default) or `"zstd"` |
| `Format` | `string` | Data format: `""` for NDJSON (default) or `"json"` for single JSON object |

#### Examples

**Basic event sending (NDJSON format):**

```go
// NDJSON: one JSON object per line
data := []byte(`{"event":"click","user":"alice"}
{"event":"view","user":"bob"}`)

err := client.SendEvents(ctx, "user_events", data, nil)
```

**Single JSON object:**

```go
data := []byte(`{"event":"purchase","amount":99.99}`)

err := client.SendEvents(ctx, "purchases", data, &tinybird.SendEventsOptions{
    Format: "json",
})
```

**With compression:**

```go
err := client.SendEvents(ctx, "events", data, &tinybird.SendEventsOptions{
    Compress: true,  // Uses gzip by default
})

// Or with zstd compression
err := client.SendEvents(ctx, "events", data, &tinybird.SendEventsOptions{
    Compress:            true,
    CompressionEncoding: "zstd",
})
```

**Wait for acknowledgment:**

```go
err := client.SendEvents(ctx, "events", data, &tinybird.SendEventsOptions{
    Wait: true,
})
```

---

### CallEndpoint

Query a Tinybird pipe endpoint.

```go
func (c *Client) CallEndpoint(
    ctx context.Context,
    endpoint string,
    params map[string]string,
    result interface{},
) error
```

#### Example

```go
params := map[string]string{
    "start_date": "2024-01-01",
    "end_date":   "2024-12-31",
    "limit":      "100",
}

err := client.CallEndpoint(ctx, "analytics_endpoint", params, nil)
```

---

### Analyze

Analyze data to get a recommended schema and preview.

```go
func (c *Client) Analyze(
    ctx context.Context,
    input interface{},
) (error, *AnalyzeResponse)
```

The `input` parameter accepts:
- `[]byte`: Raw file data to upload
- `string`: URL to a remote file

#### Examples

**Analyze local data:**

```go
data := []byte(`{"name":"Alice","age":30}
{"name":"Bob","age":25}`)

err, response := client.Analyze(ctx, data)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Recommended schema:", response.Analysis.Schema)
for _, col := range response.Analysis.Columns {
    fmt.Printf("Column: %s, Type: %s\n", col.Name, col.RecommendedType)
}
```

**Analyze remote file:**

```go
err, response := client.Analyze(ctx, "https://example.com/data.csv")
```

#### AnalyzeResponse

```go
type AnalyzeResponse struct {
    Analysis Analysis
    Preview  Preview
}

type Analysis struct {
    Columns []ColumnAnalysis  // Detected columns with recommended types
    Schema  string            // Ready-to-use schema definition
}

type ColumnAnalysis struct {
    Path            string  // JSON path to the column
    RecommendedType string  // Recommended data type
    PresentPct      int     // Percentage of non-null values
    Name            string  // Recommended column name
}

type Preview struct {
    Meta  []FieldMeta                 // Column metadata
    Data  []map[string][]interface{}  // Sample data rows
    Rows  int                         // Number of rows in preview
    Stats Statistics                  // Query statistics
}
```

## Response Types

### WriteResponse

Returned by the Events API (internal use):

```go
type WriteResponse struct {
    SuccessfulRows  int  // Number of rows successfully written
    QuarantinedRows int  // Number of rows quarantined due to errors
}
```

### EndpointResponse

Returned by pipe endpoints:

```go
type EndpointResponse struct {
    Meta            []FieldMeta                 // Column metadata
    Data            []map[string][]interface{}  // Result data
    Rows            int                         // Number of rows returned
    RowsBeforeLimit int                         // Total rows before LIMIT
    Stats           Statistics                  // Query statistics
}

type FieldMeta struct {
    Name string
    Type string
}

type Statistics struct {
    Elapsed   float64  // Query execution time in seconds
    RowsRead  int      // Number of rows read
    BytesRead int      // Number of bytes read
}
```

## Error Handling

The client returns errors for:
- Network failures (with automatic retry for 5xx and 429 status codes)
- Invalid requests (4xx status codes)
- Compression failures
- Unsupported input types

```go
err := client.SendEvents(ctx, "datasource", data, nil)
if err != nil {
    // Error message includes HTTP status and response body
    log.Printf("Failed to send events: %v", err)
}
```

## Testing

The package includes a `MockHttpClient` for testing:

```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    tinybird "github.com/NOLLYWOOD-COM/tinybird"
)

func TestMyFunction(t *testing.T) {
    mockClient := tinybird.NewMockHttpClient()

    mockClient.On("PostRaw",
        mock.Anything,
        mock.Anything,
        mock.Anything,
        mock.Anything,
        mock.Anything,
        mock.Anything,
    ).Return(nil)

    options := tinybird.DefaultClientOptions()
    client := tinybird.NewClient(options, mockClient)

    // Use client in tests...

    mockClient.AssertExpectations(t)
}
```

## License

See [LICENSE](LICENSE) for details.
