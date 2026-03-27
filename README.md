# Weiyun API V3 (MCP Wrapper)

**English** | [中文](README_zh.md)

[![CI Build and Test](https://github.com/LiuZixia/weiyun-api-v3/actions/workflows/ci.yml/badge.svg)](https://github.com/LiuZixia/weiyun-api-v3/actions/workflows/ci.yml)

> **Note**: Integration tests automatically isolate their operations within a dedicated `/CI` folder at the root of your Weiyun drive to prevent polluting personal files. It strictly validates file transfers using specific fixtures (e.g., `tencent-weiyun.zip`, `test_{language}_upload_{timestamp}`).

[](https://opensource.org/licenses/MIT)
A comprehensive, multi-language implementation of the **Tencent Weiyun V3 MCP API**. This library allows developers to integrate Weiyun cloud storage into applications, AI agents, and automated workflows.

## 🌟 Features

  * **Complete File Management**: List, batch download, batch delete, and generate sharing links.
  * **Advanced Upload Protocol**: Full implementation of the two-stage (Pre-upload + Chunked) FTN protocol.
  * **Unique SHA1 Logic**: Includes the specialized "Little-Endian SHA1 Register State" extraction required by Weiyun servers.
  * **Multi-Language Support**: Native implementations for Python, Go, PHP, and Shell.
  * **AI-Ready**: Fully compatible with the Model Context Protocol (MCP) for LLM integration.

-----

## 🛠 Architecture: The Upload Secret

The most critical part of this API is the `weiyun.upload` tool. Unlike standard cloud providers, Weiyun requires:

1.  **Intermediate Blocks**: SHA1 internal registers ($h0, h1, h2, h3, h4$) exported in **little-endian** hex format after processing each 512KB chunk.
2.  **Final Block**: A standard big-endian SHA1 hexdigest of the entire file.
3.  **Check SHA**: A specific intermediate state calculated before the final 128-byte footer.

*(Note: Replace with a sequence diagram of the Weiyun Upload Handshake if generating visual docs)*

-----

## 🚀 Getting Started

### Prerequisites

You must obtain an `mcp_token` from the [Weiyun Authorization Page](https://www.weiyun.com/act/openclaw).

### Environment Variables

```bash
export WEIYUN_MCP_TOKEN="your_token_here"
export WEIYUN_MCP_URL="https://www.weiyun.com/api/v3/mcpserver"
```

-----

## 📦 Language Implementations

### 1\. Python (Reference Implementation)

The Python version uses a custom `SHA1` class to extract internal registers that `hashlib` hides.

```python
from weiyun_api import WeiyunClient

client = WeiyunClient(token="your_token")
# Simple upload with auto-chunking and hashing
client.upload("./my_file.zip", pdir_key="root_key")
```

### 2\. Go (High Performance)

Uses a modified version of `crypto/sha1` to access the underlying `digest` struct via `reflect`.

```go
import "github.com/youruser/weiyun-api-v3/go"

func main() {
    api := weiyun.New("TOKEN")
    list, _ := api.ListFiles(50, 0)
}
```

### 3\. PHP

Ideal for web-based file managers. Implements the register rotation in pure PHP to handle 32-bit unsigned integers correctly.

```php
$weiyun = new Weiyun\Client($token);
$link = $weiyun->getDownloadLink($fileId, $pdirKey);
```

### 4\. Shell (CLI)

A wrapper around `curl` and `mcporter` for DevOps automation.

```bash
./setup.sh
# List top 10 files
mcporter call --server weiyun --tool weiyun.list limit=10
```

-----

## 📑 API Reference

| Tool | Description | Key Parameters |
| :--- | :--- | :--- |
| `weiyun.list` | Query directory contents | `limit`, `offset`, `dir_key` |
| `weiyun.download` | Get HTTPS download links | `items` (file\_id + pdir\_key) |
| `weiyun.upload` | Two-phase file upload | `file_sha`, `block_sha_list`, `file_data` |
| `weiyun.delete` | Batch delete files/folders | `file_list`, `delete_completely` |
| `weiyun.gen_share_link` | Create public share links | `file_list`, `share_name` |

-----

## ⚠️ Important Implementation Notes

1.  **pdir\_key Management**: When calling `download`, `delete`, or `gen_share_link`, always use the **top-level** `pdir_key` returned by the `weiyun.list` response. The `pdir_key` found within individual file objects may be empty.
2.  **Download Cookies**: The `weiyun.download` tool returns a URL and a Cookie. You **must** pass the cookie (e.g., `FTN5K=...`) in your GET request to avoid 403 Forbidden errors.
3.  **Rate Limiting**: Error code `117401` indicates your daily quota is exhausted.

## 🤝 Contributing

Contributions are welcome\! Specifically, we are looking for:

  * Refined Go implementations for faster SHA1 register extraction.
  * Node.js/TypeScript port.
  * Improved documentation for the `check_data` Base64 padding.

## 📜 License

MIT License. See `LICENSE` for details.

-----

Would you like me to generate the **Go** or **PHP** boilerplate code for the specialized SHA1 register extraction to include in this library?