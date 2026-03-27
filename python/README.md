# Weiyun API V3 - Python Client

This directory contains the Python reference implementation for the Tencent Weiyun V3 MCP API.

## Requirements
- Python 3.7+
- `requests`

## Installation
Ensure you have the requests library installed:
```bash
pip install requests
```

## Quick Start

```python
from weiyun_api import WeiyunClient

# Initialize the client with your token
client = WeiyunClient(token="your_mcp_token_here")

# 1. List Files in a Directory
# Returns top 50 files/directories at the root
response = client.list_files(limit=50, offset=0)
print(response)

# 2. Upload a File
# Handles the FTN protocol's two-stage hashing automatically
upload_res = client.upload(file_path="./example.txt", pdir_key="optional_dir_key")
print(f"Uploaded! File ID: {upload_res['file_id']}")

# 3. Get Download Link
# Supply file_id and its parent directory key (pdir_key)
download_res = client.download([{"file_id": "file_123", "pdir_key": "dir_456"}])
print("Download URL:", download_res['items'][0]['https_download_url'])

# 4. Generate Sharing Link
share_res = client.gen_share_link(file_list=[{"file_id": "file_123", "pdir_key": "dir_456"}])
print("Share Link:", share_res['short_url'])

# 5. Delete Files
del_res = client.delete(file_list=[{"file_id": "file_123", "pdir_key": "dir_456"}], delete_completely=False)
print("Deleted.")
```

## Class Reference: `WeiyunClient`

- `__init__(token: str, mcp_url: str = "...")`: Initializes the client.
- `list_files(limit: int = 50, get_type: int = 0, offset: int = 0, dir_key: str = None, pdir_key: str = None)`: Queries files and folders.
- `upload(file_path: str, pdir_key: str = None, max_rounds: int = 50)`: Streams files securely via Weiyun's 2-phase protocol with internal small-endian SHA1 parsing.
- `download(items: list)`: Grabs HTTPS download links and mandatory cookies.
- `delete(file_list: list = None, dir_list: list = None, delete_completely: bool = False)`: Moves files or folders to trash or deletes them entirely.
- `gen_share_link(file_list: list = None, dir_list: list = None, share_name: str = None)`: Generates short share URLs.
