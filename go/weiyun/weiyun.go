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

type Client struct {
	Token  string
	McpURL string
}

func New(token string) *Client {
	return &Client{
		Token:  token,
		McpURL: "https://www.weiyun.com/api/v3/mcpserver",
	}
}

// Extract internal state of crypto/sha1 digest via reflect and unsafe
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

func GetInternalState(s reflect.Value) string {
	return getSha1State(s)
}

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

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) ListFiles(limit int, offset int) (map[string]interface{}, error) {
	return c.Call("weiyun.list", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	})
}

// Upload params logic 
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
