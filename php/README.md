# Weiyun API V3 - PHP Client

This directory provides the Web-centric PHP implementation of the Tencent Weiyun V3 MCP API toolkit. 

Due to integers differing in length across 32-bit and 64-bit PHP installations, our solution carries a standalone `SHA1` rotating calculator that explicitly masks and handles strict integer overflow matching (`& 0xFFFFFFFF`) natively in PHP!

## Requirements
- PHP 7.4 / 8.0+
- `PHPUnit` (if you intend to run test coverage)

## Quick Start
You do not need Composer to use this locally, although integrating it into a composer package is trivial.

```php
<?php
require_once __DIR__ . '/src/Weiyun/Client.php';
use Weiyun\Client;

$token = "your_mcp_token_here";
$client = new Client($token);

// 1. List Files
$listRes = $client->listFiles(10, 0);
print_r($listRes);

// 2. Obtain an HTTPS Download Link
// Keep in mind when generating share or download links 
// you MUST have the root folder key, not just the file ID!
$downloadUrlObject = $client->getDownloadLink("file_123", "dir_456");
echo "URL: " . $downloadUrlObject['items'][0]['https_download_url'] . "\n";

// 3. Perform the Upload Hashing Phase
// Because the server does not accept final hashes, it requires partial chunk hashes.
$uploadBlockInfo = $client->calcUploadParams("/tmp/my_file.zip");
echo "Chunk Count: " . count($uploadBlockInfo['block_sha_list']) . "\n";
print_r($uploadBlockInfo);

// 4. Custom Calling (For deletion or sharing bindings)
$response = $client->call("weiyun.gen_share_link", [
    "file_list" => [
        ["file_id" => "file_123", "pdir_key" => "dir_456"]
    ]
]);
print_r($response);
```

## Class Reference

### `Weiyun\Client`
* `__construct($token, $mcpUrl)`: Connects client references via POST stream contexts.
* `call($toolName, $arguments)`: Submits arbitrary RPC `tools/call`.
* `listFiles($limit, $offset)`: Fetches directory contents.
* `getDownloadLink($fileId, $pdirKey)`: Binds items to secure HTTP resources.
* `calcUploadParams($filePath)`: Feeds large files cleanly into our integrated PHP custom un-finalized SHA extraction method bridging the FTN constraints. 

### `Weiyun\SHA1`
*The internal rotation engine. Used strictly inside `calcUploadParams()` for processing the standard 524288 byte hashing requirements in pure PHP syntax.*
- `get_state()`: Packs `V5` internal Little-Endian integers into a clean hex string representing the H0-H4 buffers. 
- `hexdigest()`: Fallback standard 64 byte 0x80 padding calculator matching equivalent systems.
