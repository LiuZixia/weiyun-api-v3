package weiyun

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := New("TEST_TOKEN")
	if client.Token != "TEST_TOKEN" {
		t.Errorf("Expected token TEST_TOKEN, got %s", client.Token)
	}
	if client.McpURL != "https://www.weiyun.com/api/v3/mcpserver" {
		t.Errorf("Expected URL https://www.weiyun.com/api/v3/mcpserver, got %s", client.McpURL)
	}
}

func TestListFiles(t *testing.T) {
	// Simple mock test idea - since no httptest setup, just checking compilation
	client := New("TEST_TOKEN")
	if client == nil {
		t.Errorf("Client should not be nil")
	}
}

func TestCalcUploadParams(t *testing.T) {
	// Assuming test on empty or missing file fails gracefully
	_, err := CalcUploadParams("nonexistent.txt")
	if err == nil {
		t.Errorf("Expected error for nonexistent file")
	}
}
