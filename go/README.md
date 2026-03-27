# Weiyun API V3 - Go Client

This directory contains the Go implementation of the Tencent Weiyun V3 MCP API. The client uses `reflect` and `unsafe.Pointer` to extract the `crypto/sha1` internal registers required by the Weiyun upload protocol.

## Requirements
- Go 1.20+

## Installation
```bash
go mod tidy
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/youruser/weiyun-api-v3/go/weiyun"
)

func main() {
    api := weiyun.New("your_mcp_token_here")

    // 1. List Files
    listRes, err := api.ListFiles(50, 0)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("File List:", listRes)

    // 2. Upload a File (two-phase FTN protocol, auto-chunked)
    upRes, err := api.Upload("./my_file.zip", "" /* root */, 50)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Uploaded! File ID:", upRes["file_id"])

    // 3. Get HTTPS Download Links
    dlRes, err := api.Download([]map[string]interface{}{
        {"file_id": "file_123", "pdir_key": "dir_456"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Download Response:", dlRes)

    // 4. Generate a Share Link
    shareRes, err := api.GenShareLink(
        []map[string]interface{}{{"file_id": "file_123", "pdir_key": "dir_456"}},
        nil,
        "My Share",
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Share:", shareRes["short_url"])

    // 5. Delete Files
    delRes, err := api.Delete(
        []map[string]interface{}{{"file_id": "file_123", "pdir_key": "dir_456"}},
        nil,
        false, // false = move to trash, true = permanent delete
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Deleted:", delRes)
}
```

## Running Tests

### Unit Tests
```bash
go test ./weiyun/... -v
```

### Integration Tests

Requires a folder named `CI` at the root of your Weiyun drive.

```bash
export WEIYUN_MCP_TOKEN="your_token_here"
go test -tags integration ./weiyun/... -v
```

## Package Reference

| Function / Method | Description |
|---|---|
| `New(token string) *Client` | Create a client pointing to the official MCP endpoint |
| `(c *Client) Call(tool string, args map[string]interface{})` | Generic JSON-RPC `tools/call` |
| `(c *Client) ListFiles(limit int, offset int)` | List files and directories |
| `(c *Client) Download(items []map[string]interface{})` | Get HTTPS download links |
| `(c *Client) Delete(fileList, dirList []map[string]interface{}, deleteCompletely bool)` | Delete files/folders |
| `(c *Client) GenShareLink(fileList, dirList []map[string]interface{}, shareName string)` | Generate a public share link |
| `CalcUploadParams(filePath string)` | Compute chunked SHA1 states + MD5 for upload |
| `(c *Client) Upload(filePath, pdirKey string, maxRounds int)` | Full two-phase FTN upload with auto-retry |
