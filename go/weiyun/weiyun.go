package weiyun

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"time"
	"unsafe"
)

// Client holds the Weiyun MCP API credentials.
type Client struct {
	Token  string
	McpURL string
}

// New creates a new Client with default endpoint.
func New(token string) *Client {
	return &Client{
		Token:  token,
		McpURL: "https://www.weiyun.com/api/v3/mcpserver",
	}
}

// getSha1State extracts the internal little-endian h0-h4 state of a crypto/sha1 hash
// via reflect and unsafe, as required by the Weiyun upload protocol.
func getSha1State(h reflect.Value) string {
	d := h.Elem()
	hField := d.FieldByName("h")
	ptr := unsafe.Pointer(hField.UnsafeAddr())
	hArr := *(*[5]uint32)(ptr)

	h0, h1, h2, h3, h4 := hArr[0], hArr[1], hArr[2], hArr[3], hArr[4]

	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x",
		byte(h0), byte(h0>>8), byte(h0>>16), byte(h0>>24),
		byte(h1), byte(h1>>8), byte(h1>>16), byte(h1>>24),
		byte(h2), byte(h2>>8), byte(h2>>16), byte(h2>>24),
		byte(h3), byte(h3>>8), byte(h3>>16), byte(h3>>24),
		byte(h4), byte(h4>>8), byte(h4>>16), byte(h4>>24),
	)
}

// GetInternalState is an exported helper for tests.
func GetInternalState(s reflect.Value) string {
	return getSha1State(s)
}

// parseResult extracts the text content from a JSON-RPC response.
func parseResult(result map[string]interface{}) (map[string]interface{}, error) {
	rr, ok := result["result"].(map[string]interface{})
	if !ok {
		return result, nil
	}
	content, ok := rr["content"].([]interface{})
	if !ok {
		return result, nil
	}
	for _, item := range content {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] == "text" {
			var out map[string]interface{}
			if err := json.Unmarshal([]byte(m["text"].(string)), &out); err != nil {
				return nil, err
			}
			return out, nil
		}
	}
	return result, nil
}

// Call makes a JSON-RPC tools/call request to the MCP server.
func (c *Client) Call(tool string, args map[string]interface{}) (map[string]interface{}, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      tool,
			"arguments": args,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.McpURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("WyHeader", "mcp_token="+c.Token)

	httpClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return parseResult(raw)
}

// ListFiles lists files and directories.
func (c *Client) ListFiles(limit int, offset int) (map[string]interface{}, error) {
	return c.Call("weiyun.list", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})
}

// Download retrieves HTTPS download links for a list of items.
// Each item should be map[string]interface{}{"file_id": "...", "pdir_key": "..."}.
func (c *Client) Download(items []map[string]interface{}) (map[string]interface{}, error) {
	// Convert to []interface{} for JSON serialisation
	its := make([]interface{}, len(items))
	for i, v := range items {
		its[i] = v
	}
	return c.Call("weiyun.download", map[string]interface{}{
		"items": its,
	})
}

// Delete permanently removes or trashes files/directories.
func (c *Client) Delete(
	fileList []map[string]interface{},
	dirList []map[string]interface{},
	deleteCompletely bool,
) (map[string]interface{}, error) {
	args := map[string]interface{}{
		"delete_completely": deleteCompletely,
	}
	if len(fileList) > 0 {
		fl := make([]interface{}, len(fileList))
		for i, v := range fileList {
			fl[i] = v
		}
		args["file_list"] = fl
	}
	if len(dirList) > 0 {
		dl := make([]interface{}, len(dirList))
		for i, v := range dirList {
			dl[i] = v
		}
		args["dir_list"] = dl
	}
	return c.Call("weiyun.delete", args)
}

// GenShareLink generates a public share link for files and/or directories.
func (c *Client) GenShareLink(
	fileList []map[string]interface{},
	dirList []map[string]interface{},
	shareName string,
) (map[string]interface{}, error) {
	args := map[string]interface{}{}
	if len(fileList) > 0 {
		fl := make([]interface{}, len(fileList))
		for i, v := range fileList {
			fl[i] = v
		}
		args["file_list"] = fl
	}
	if len(dirList) > 0 {
		dl := make([]interface{}, len(dirList))
		for i, v := range dirList {
			dl[i] = v
		}
		args["dir_list"] = dl
	}
	if shareName != "" {
		args["share_name"] = shareName
	}
	return c.Call("weiyun.gen_share_link", args)
}

// CalcUploadParams computes all SHA1 / MD5 parameters required by the Weiyun upload protocol.
func CalcUploadParams(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()
	blockSize := int64(524288)

	lastBlockSize := fileSize % blockSize
	if lastBlockSize == 0 {
		lastBlockSize = blockSize
	}
	checkBlockSize := lastBlockSize % 128
	if checkBlockSize == 0 {
		checkBlockSize = 128
	}
	beforeBlockSize := fileSize - lastBlockSize

	h := sha1.New()
	v := reflect.ValueOf(h)

	m := md5.New()
	var blockShaList []string

	for offset := int64(0); offset < beforeBlockSize; offset += blockSize {
		buf := make([]byte, blockSize)
		io.ReadFull(file, buf)
		h.Write(buf)
		m.Write(buf)
		blockShaList = append(blockShaList, getSha1State(v))
	}

	betweenData := make([]byte, lastBlockSize-checkBlockSize)
	io.ReadFull(file, betweenData)
	h.Write(betweenData)
	m.Write(betweenData)
	checkSha := getSha1State(v)

	checkDataBytes := make([]byte, checkBlockSize)
	io.ReadFull(file, checkDataBytes)
	h.Write(checkDataBytes)
	m.Write(checkDataBytes)

	fileSha := fmt.Sprintf("%x", h.Sum(nil))
	checkData := base64.StdEncoding.EncodeToString(checkDataBytes)
	fileMd5 := fmt.Sprintf("%x", m.Sum(nil))

	blockShaList = append(blockShaList, fileSha)

	return map[string]interface{}{
		"filename":       stat.Name(),
		"file_size":      fileSize,
		"file_sha":       fileSha,
		"file_md5":       fileMd5,
		"block_sha_list": blockShaList,
		"check_sha":      checkSha,
		"check_data":     checkData,
	}, nil
}

// Upload uploads a file using the Weiyun two-phase FTN protocol.
// pdirKey is the target directory key (empty string = root).
// maxRounds is the maximum number of upload loop iterations.
func (c *Client) Upload(filePath string, pdirKey string, maxRounds int) (map[string]interface{}, error) {
	params, err := CalcUploadParams(filePath)
	if err != nil {
		return nil, err
	}

	fileSize := params["file_size"].(int64)
	filename := params["filename"].(string)

	preArgs := map[string]interface{}{
		"filename":       filename,
		"file_size":      fileSize,
		"file_sha":       params["file_sha"],
		"file_md5":       params["file_md5"],
		"block_sha_list": params["block_sha_list"],
		"check_sha":      params["check_sha"],
		"check_data":     params["check_data"],
	}
	if pdirKey != "" {
		preArgs["pdir_key"] = pdirKey
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	for round := 0; round < maxRounds; round++ {
		preRsp, err := c.Call("weiyun.upload", preArgs)
		if err != nil {
			return nil, err
		}

		if errMsg, ok := preRsp["error"]; ok && errMsg != nil {
			return nil, fmt.Errorf("预上传失败: %v", errMsg)
		}

		if fe, ok := preRsp["file_exist"].(bool); ok && fe {
			return map[string]interface{}{
				"file_id":  strOrEmpty(preRsp["file_id"]),
				"filename": strOrEmpty(preRsp["filename"]),
			}, nil
		}

		chList := toSliceOfMaps(preRsp["channel_list"])
		uk := strOrEmpty(preRsp["upload_key"])
		ex := strOrEmpty(preRsp["ex"])

		// Find first channel with remaining data
		var ch map[string]interface{}
		for _, c2 := range chList {
			if toInt64(c2["len"]) > 0 {
				ch = c2
				break
			}
		}

		if ch == nil {
			state := toInt64(preRsp["upload_state"])
			if state == 2 {
				return map[string]interface{}{
					"file_id":  strOrEmpty(preRsp["file_id"]),
					"filename": strOrEmpty(preRsp["filename"]),
				}, nil
			}
			return nil, fmt.Errorf("无可上传通道，upload_state=%d", state)
		}

		offset := toInt64(ch["offset"])
		length := toInt64(ch["len"])
		channelID := toInt64(ch["id"])

		actualLen := length
		if offset+length > int64(len(fileData)) {
			actualLen = int64(len(fileData)) - offset
		}

		chunk := fileData[offset : offset+actualLen]
		chunkB64 := base64.StdEncoding.EncodeToString(chunk)

		cl := make([]interface{}, len(chList))
		for i, c2 := range chList {
			cl[i] = map[string]interface{}{
				"id":     toInt64(c2["id"]),
				"offset": toInt64(c2["offset"]),
				"len":    toInt64(c2["len"]),
			}
		}

		upRsp, err := c.Call("weiyun.upload", map[string]interface{}{
			"filename":       filename,
			"file_size":      fileSize,
			"file_sha":       params["file_sha"],
			"block_sha_list": []interface{}{},
			"check_sha":      params["check_sha"],
			"upload_key":     uk,
			"channel_list":   cl,
			"channel_id":     channelID,
			"ex":             ex,
			"file_data":      chunkB64,
		})
		if err != nil {
			return nil, err
		}

		if errMsg, ok := upRsp["error"]; ok && errMsg != nil {
			return nil, fmt.Errorf("分片上传失败: %v", errMsg)
		}

		if toInt64(upRsp["upload_state"]) == 2 {
			return map[string]interface{}{
				"file_id":  strOrEmpty(upRsp["file_id"]),
				"filename": strOrEmpty(upRsp["filename"]),
			}, nil
		}
	}

	return nil, fmt.Errorf("超过最大上传轮数 (%d)，上传未完成", maxRounds)
}

// ─── Utility helpers ──────────────────────────────────────────────────────────

func strOrEmpty(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

func toSliceOfMaps(v interface{}) []map[string]interface{} {
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if ok {
			result = append(result, m)
		}
	}
	return result
}
