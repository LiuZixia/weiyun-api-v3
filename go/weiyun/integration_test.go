//go:build integration

package weiyun

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// Integration tests hit the real Weiyun API endpoint.
// Run with: WEIYUN_MCP_TOKEN=xxx go test -tags integration ./weiyun/...
//
// Requires a folder named "CI" at the root of your Weiyun drive.

func integrationClient(t *testing.T) *Client {
	t.Helper()
	token := os.Getenv("WEIYUN_MCP_TOKEN")
	if token == "" {
		t.Fatal("WEIYUN_MCP_TOKEN is not set")
	}
	return New(token)
}

func getCiDirKeyIntegration(t *testing.T, c *Client) (ciDirKey, rootPdir string) {
	t.Helper()
	offset := 0
	limit := 50
	for {
		res, err := c.ListFiles(limit, offset)
		if err != nil {
			t.Fatalf("ListFiles failed: %v", err)
		}
		if p, ok := res["pdir_key"].(string); ok && rootPdir == "" {
			rootPdir = p
		}
		dirList := toSliceOfMaps(res["dir_list"])
		for _, d := range dirList {
			if d["dir_name"] == "CI" {
				ciDirKey = d["dir_key"].(string)
				t.Logf("✅ Found /CI directory: %s", ciDirKey)
				return
			}
		}
		finish, _ := res["finish_flag"].(bool)
		if finish {
			break
		}
		offset += limit
	}
	t.Fatal("⚠️  /CI directory not found. Please create a folder named 'CI' in the root of your Weiyun drive.")
	return
}

func TestIntegrationListCiDir(t *testing.T) {
	c := integrationClient(t)
	ciDirKey, rootPdir := getCiDirKeyIntegration(t, c)

	// List files inside /CI using dir_key and pdir_key
	res, err := c.Call("weiyun.list", map[string]interface{}{
		"limit":    10,
		"offset":   0,
		"dir_key":  ciDirKey,
		"pdir_key": rootPdir,
	})
	if err != nil {
		t.Fatalf("list CI dir failed: %v", err)
	}
	if _, ok := res["file_list"]; !ok {
		t.Error("Expected file_list in response")
	}
	if _, ok := res["dir_list"]; !ok {
		t.Error("Expected dir_list in response")
	}
	t.Log("✅ Successfully listed files in /CI directory.")
}

func TestIntegrationUploadFile(t *testing.T) {
	c := integrationClient(t)
	ciDirKey, _ := getCiDirKeyIntegration(t, c)

	filename := fmt.Sprintf("test_go_upload_%d", time.Now().Unix())
	if err := os.WriteFile(filename, []byte("Hello from CI Upload Test!"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	res, err := c.Upload(filename, ciDirKey, 50)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	if _, ok := res["file_id"]; !ok {
		t.Error("Expected file_id in upload response")
	}
	t.Logf("✅ Successfully uploaded %s to /CI.", filename)
}

func TestIntegrationDownloadFile(t *testing.T) {
	c := integrationClient(t)
	ciDirKey, _ := getCiDirKeyIntegration(t, c)

	filename := "tencent-weiyun.zip"
	if err := os.WriteFile(filename, []byte("Dummy ZIP content for download testing"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	upRes, err := c.Upload(filename, ciDirKey, 50)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	fileID := upRes["file_id"].(string)

	dlRes, err := c.Download([]map[string]interface{}{
		{"file_id": fileID, "pdir_key": ciDirKey},
	})
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	items := toSliceOfMaps(dlRes["items"])
	if len(items) == 0 {
		t.Fatal("Expected at least one item in download response")
	}
	if _, ok := items[0]["https_download_url"]; !ok {
		t.Error("Expected https_download_url in download item")
	}
	t.Logf("✅ Successfully obtained real download link for %s.", filename)
}

func TestIntegrationDeleteFile(t *testing.T) {
	c := integrationClient(t)
	ciDirKey, _ := getCiDirKeyIntegration(t, c)

	filename := fmt.Sprintf("test_go_delete_%d.txt", time.Now().Unix())
	if err := os.WriteFile(filename, []byte("Will automatically delete me!"), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	upRes, err := c.Upload(filename, ciDirKey, 50)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}
	fileID := upRes["file_id"].(string)

	delRes, err := c.Delete(
		[]map[string]interface{}{{"file_id": fileID, "pdir_key": ciDirKey}},
		nil,
		true,
	)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	cnt := toInt64(delRes["freed_index_cnt"])
	if cnt < 1 {
		t.Errorf("Expected freed_index_cnt >= 1, got %d", cnt)
	}
	t.Logf("✅ Successfully deleted %s from /CI.", filename)
}
