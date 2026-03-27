package weiyun

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// mockMCPResponse builds a minimal JSON-RPC response wrapping the given data.
func mockMCPResponse(t *testing.T, data map[string]interface{}) []byte {
	t.Helper()
	text, _ := json.Marshal(data)
	outer := map[string]interface{}{
		"result": map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{"type": "text", "text": string(text)},
			},
		},
	}
	b, _ := json.Marshal(outer)
	return b
}

// newTestClient creates a Client pointing at a test server and a cleanup func.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := &Client{Token: "TEST_TOKEN", McpURL: srv.URL}
	return c, srv.Close
}

// ─── Unit Tests ───────────────────────────────────────────────────────────────

func TestNewClient(t *testing.T) {
	client := New("TEST_TOKEN")
	if client.Token != "TEST_TOKEN" {
		t.Errorf("Expected token TEST_TOKEN, got %s", client.Token)
	}
	if client.McpURL != "https://www.weiyun.com/api/v3/mcpserver" {
		t.Errorf("Unexpected McpURL: %s", client.McpURL)
	}
}

func TestCalcUploadParams(t *testing.T) {
	_, err := CalcUploadParams("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCalcUploadParamsSmallFile(t *testing.T) {
	f, err := os.CreateTemp("", "weiyun_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("Hello Weiyun!")
	f.Close()

	params, err := CalcUploadParams(f.Name())
	if err != nil {
		t.Fatalf("CalcUploadParams failed: %v", err)
	}
	for _, key := range []string{"filename", "file_size", "file_sha", "file_md5", "block_sha_list", "check_sha", "check_data"} {
		if _, ok := params[key]; !ok {
			t.Errorf("Missing key: %s", key)
		}
	}
}

func TestListFiles(t *testing.T) {
	payload := map[string]interface{}{
		"file_list": []interface{}{map[string]interface{}{"file_id": "123"}},
		"dir_list":  []interface{}{},
	}
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	res, err := client.ListFiles(10, 0)
	if err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}
	if _, ok := res["file_list"]; !ok {
		t.Error("Expected file_list in response")
	}
}

func TestDownload(t *testing.T) {
	payload := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"file_id": "123", "https_download_url": "https://example.com/file"},
		},
	}
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	res, err := client.Download([]map[string]interface{}{
		{"file_id": "123", "pdir_key": "dir_abc"},
	})
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if _, ok := res["items"]; !ok {
		t.Error("Expected items in response")
	}
}

func TestDelete(t *testing.T) {
	payload := map[string]interface{}{"freed_index_cnt": float64(1)}
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	res, err := client.Delete(
		[]map[string]interface{}{{"file_id": "123", "pdir_key": "dir_abc"}},
		nil,
		true,
	)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if cnt, ok := res["freed_index_cnt"]; !ok || toInt64(cnt) < 1 {
		t.Error("Expected freed_index_cnt >= 1")
	}
}

func TestGenShareLink(t *testing.T) {
	payload := map[string]interface{}{"short_url": "https://share.weiyun.com/abc123"}
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	res, err := client.GenShareLink(
		[]map[string]interface{}{{"file_id": "123", "pdir_key": "dir_abc"}},
		nil,
		"My Share",
	)
	if err != nil {
		t.Fatalf("GenShareLink error: %v", err)
	}
	if _, ok := res["short_url"]; !ok {
		t.Error("Expected short_url in response")
	}
}

func TestUploadFileExist(t *testing.T) {
	payload := map[string]interface{}{
		"file_exist": true,
		"file_id":    "existing_123",
		"filename":   "test.txt",
	}
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	f, _ := os.CreateTemp("", "weiyun_upload_*.txt")
	defer os.Remove(f.Name())
	f.WriteString("Hello Weiyun!")
	f.Close()

	res, err := client.Upload(f.Name(), "", 10)
	if err != nil {
		t.Fatalf("Upload error: %v", err)
	}
	if res["file_id"] != "existing_123" {
		t.Errorf("Expected file_id existing_123, got %v", res["file_id"])
	}
}

func TestUploadChunkedCompletes(t *testing.T) {
	f, _ := os.CreateTemp("", "weiyun_upload_*.txt")
	defer os.Remove(f.Name())
	f.WriteString("Hello Weiyun chunked!")
	f.Close()

	fi, _ := os.Stat(f.Name())
	fileSize := fi.Size()

	call := 0
	client, cleanup := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		call++
		var payload map[string]interface{}
		if call == 1 {
			// Pre-upload response: returns a channel to fill
			payload = map[string]interface{}{
				"file_exist": false,
				"upload_key": "uk_abc",
				"ex":         "",
				"channel_list": []interface{}{
					map[string]interface{}{"id": float64(1), "offset": float64(0), "len": float64(fileSize)},
				},
				"upload_state": float64(1),
			}
		} else {
			// Chunk upload response: done
			payload = map[string]interface{}{
				"upload_state": float64(2),
				"file_id":      "new_file_456",
				"filename":     f.Name(),
			}
		}
		w.Write(mockMCPResponse(t, payload))
	})
	defer cleanup()

	res, err := client.Upload(f.Name(), "", 10)
	if err != nil {
		t.Fatalf("Upload error: %v", err)
	}
	if res["file_id"] != "new_file_456" {
		t.Errorf("Expected file_id new_file_456, got %v", res["file_id"])
	}
}
