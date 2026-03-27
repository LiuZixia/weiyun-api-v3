# Weiyun API V3 - PHP Client

This directory provides the PHP implementation of the Tencent Weiyun V3 MCP API toolkit.

Due to integer differences across 32-bit and 64-bit PHP installations, the solution carries a standalone `SHA1` class that explicitly masks integer overflow with `& 0xFFFFFFFF`, ensuring byte-accurate little-endian register state extraction required by the Weiyun upload protocol.

## Requirements
- PHP 7.4 / 8.0+
- `PHPUnit` (optional, for running unit tests)

## Quick Start

No Composer required for basic usage.

```php
<?php
require_once __DIR__ . '/src/Weiyun/Client.php';
use Weiyun\Client;

$token = "your_mcp_token_here";
$client = new Client($token);

// 1. List Files
$listRes = $client->listFiles(50, 0, 0);
print_r($listRes);

// 2. Upload a File (two-phase FTN protocol, auto-chunked)
$upRes = $client->upload("/tmp/my_file.zip", $pdirKey = null);
echo "Uploaded! File ID: " . $upRes['file_id'] . "\n";

// 3. Get HTTPS Download Links
$dlRes = $client->download([["file_id" => "file_123", "pdir_key" => "dir_456"]]);
echo "URL: " . $dlRes['items'][0]['https_download_url'] . "\n";

// 4. Generate a Share Link
$shareRes = $client->genShareLink(
    [["file_id" => "file_123", "pdir_key" => "dir_456"]],
    null,
    "My Share"
);
echo "Share: " . $shareRes['short_url'] . "\n";

// 5. Delete Files
$delRes = $client->delete(
    [["file_id" => "file_123", "pdir_key" => "dir_456"]],
    null,
    false  // false = move to trash, true = permanent delete
);
print_r($delRes);
```

## Running Tests

### Unit Tests (PHPUnit)
```bash
./vendor/bin/phpunit tests/
```

Or standalone (no Composer):
```bash
php tests/ClientTest.php
```

### Integration Tests

Requires a folder named `CI` at the root of your Weiyun drive.

```bash
export WEIYUN_MCP_TOKEN="your_token_here"
php integration_tests.php --test all
# Options: list | upload | download | delete | all
```

## Class Reference

### `Weiyun\Client`

| Method | Description |
|---|---|
| `__construct($token, $mcpUrl)` | Initializes the client |
| `call($toolName, $arguments)` | Arbitrary JSON-RPC `tools/call` |
| `listFiles($limit, $getType, $offset, $dirKey, $pdirKey)` | List files and directories |
| `download($items)` | Get HTTPS download links for a list of `["file_id", "pdir_key"]` items |
| `delete($fileList, $dirList, $deleteCompletely)` | Delete files/folders (trash or permanent) |
| `genShareLink($fileList, $dirList, $shareName)` | Generate a public share link |
| `calcUploadParams($filePath)` | Compute chunked SHA1 states + MD5 for upload |
| `upload($filePath, $pdirKey, $maxRounds)` | Full two-phase FTN upload with auto-retry |

### `Weiyun\SHA1`

*Internal rotation engine used by `calcUploadParams()` for the 512KB block hashing.*

| Method | Description |
|---|---|
| `update($data)` | Feed data into the SHA1 state |
| `get_state()` | Export little-endian H0–H4 as hex (intermediate state) |
| `hexdigest()` | Standard big-endian SHA1 hex string of full input |
