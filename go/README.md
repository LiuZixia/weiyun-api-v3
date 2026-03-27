# Weiyun API V3 - Go Client

This directory contains the high-performance Go implementation for the Tencent Weiyun V3 MCP API. The client circumvents standard language access limits by using `reflect` and `unsafe.Pointer` to extract the `crypto/sha1` internal structures required for the Weiyun upload protocol.

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
	// Initialize a client instance
	api := weiyun.New("your_mcp_token_here")

	// 1. List Files
	// Retrieve a map of the top 10 items at the root
	listResp, err := api.ListFiles(10, 0)
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}
	fmt.Println("File List Response:", listResp)

	// 2. Upload Pre-Calculation 
	// The client auto-calculates piece intervals and exposes unfinalized SHA pieces
	params, err := weiyun.CalcUploadParams("./test.txt")
	if err != nil {
		log.Fatalf("Checksum extraction error: %v", err)
	}
	
	fmt.Printf("Custom SHA Internal Blocks Checksum Setup:\n%#v\n", params)

	// 3. Generic MCP Caller (Used for building Download/Delete abstractions)
	args := map[string]interface{}{
		"items": []map[string]string{
			{"file_id": "file_123", "pdir_key": "dir_456"},
		},
	}
	
	downloadResp, err := api.Call("weiyun.download", args)
	fmt.Println(downloadResp)
}
```

## Package Functions
* `New(token string) *Client`: Instantiates the API wrapper pointing to the official `api/v3/mcpserver` endpoints.
* `(c *Client) Call(tool string, args map[string]interface{})`: The core abstract caller to bridge local environments into MCP interactions.
* `(c *Client) ListFiles(limit int, offset int)`: Quickly retrieve remote documents.
* `CalcUploadParams(filePath string)`: Evaluates files with a custom 512KB offset stream tracking SHA1 internal pointers (`h0` to `h4`) into small-endian byte formatting sequences.
