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
  * **Multi-Language Support**: Native implementations for Python, Go, and PHP.
  * **AI-Ready**: Fully compatible with the Model Context Protocol (MCP) for LLM integration.

-----

## 🛠 Architecture: The Upload Secret

The most critical part of this API is the `weiyun.upload` tool. Unlike standard cloud providers, Weiyun requires:

1.  **Intermediate Blocks**: SHA1 internal registers ($h0, h1, h2, h3, h4$) exported in **little-endian** hex format after processing each 512KB chunk.
2.  **Final Block**: A standard big-endian SHA1 hexdigest of the entire file.
3.  **Check SHA**: A specific intermediate state calculated before the final 128-byte footer.

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
client.upload("./my_file.zip", pdir_key="root_key")
res = client.download([{"file_id": "f_123", "pdir_key": "d_456"}])
client.delete(file_list=[{"file_id": "f_123", "pdir_key": "d_456"}], delete_completely=True)
client.gen_share_link(file_list=[{"file_id": "f_123", "pdir_key": "d_456"}])
```

**Tests:**
```bash
cd python
pip install requests
python test_weiyun_api.py                        # unit tests
WEIYUN_MCP_TOKEN=xxx python integration_tests.py # integration tests
```

### 2\. Go (High Performance)

Uses `reflect` and `unsafe.Pointer` to access the `crypto/sha1` internal digest struct.

```go
import "github.com/youruser/weiyun-api-v3/go/weiyun"

api := weiyun.New("TOKEN")
api.Upload("./my_file.zip", "", 50)
api.Download([]map[string]interface{}{{"file_id": "f_123", "pdir_key": "d_456"}})
api.Delete([]map[string]interface{}{{"file_id": "f_123", "pdir_key": "d_456"}}, nil, true)
api.GenShareLink([]map[string]interface{}{{"file_id": "f_123", "pdir_key": "d_456"}}, nil, "")
```

**Tests:**
```bash
cd go
go test ./weiyun/... -v                                          # unit tests
WEIYUN_MCP_TOKEN=xxx go test -tags integration ./weiyun/... -v  # integration tests
```

### 3\. PHP

Ideal for web-based file managers. Implements the register rotation in pure PHP to handle 32-bit unsigned integers correctly.

```php
$client = new Weiyun\Client($token);
$client->upload("/tmp/my_file.zip");
$client->download([["file_id" => "f_123", "pdir_key" => "d_456"]]);
$client->delete([["file_id" => "f_123", "pdir_key" => "d_456"]], null, true);
$client->genShareLink([["file_id" => "f_123", "pdir_key" => "d_456"]]);
```

**Tests:**
```bash
cd php
./vendor/bin/phpunit tests/                                        # unit tests
WEIYUN_MCP_TOKEN=xxx php integration_tests.php --test all         # integration tests
```

-----

## 📑 API Reference

| Tool | Description | Key Parameters |
| :--- | :--- | :--- |
| `weiyun.list` | Query directory contents | `limit`, `offset`, `dir_key`, `pdir_key` |
| `weiyun.download` | Get HTTPS download links | `items` (file\_id + pdir\_key) |
| `weiyun.upload` | Two-phase file upload | `file_sha`, `block_sha_list`, `file_data` |
| `weiyun.delete` | Batch delete files/folders | `file_list`, `dir_list`, `delete_completely` |
| `weiyun.gen_share_link` | Create public share links | `file_list`, `dir_list`, `share_name` |

-----

## ⚠️ Important Implementation Notes

1.  **pdir\_key Management**: When calling `download`, `delete`, or `gen_share_link`, always use the **top-level** `pdir_key` returned by the `weiyun.list` response. The `pdir_key` found within individual file objects may be empty.
2.  **Download Cookies**: The `weiyun.download` tool returns a URL and a Cookie. You **must** pass the cookie (e.g., `FTN5K=...`) in your GET request to avoid 403 Forbidden errors.
3.  **Rate Limiting**: Error code `117401` indicates your daily quota is exhausted.

## 🤝 Contributing

Contributions are welcome\! Specifically, we are looking for:

  * Node.js/TypeScript port.
  * Improved documentation for the `check_data` Base64 padding.

## 📜 License

MIT License. See `LICENSE` for details.