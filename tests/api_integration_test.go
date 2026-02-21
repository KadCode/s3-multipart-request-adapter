package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Use global s3Client and testBucket from TestMain
var baseURL = "https://localhost:8080/ContentServer/ContentServer.dll"

// TestUploadDownloadDelete performs full integration test: upload, download, delete
func TestUploadDownloadDelete(t *testing.T) {
	//defer cleanupTestBucket() // optional extra cleanup
	docID := "TEST-INTEGRATION-FILE"

	// ---- Upload ----
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filePath := "test.txt"
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create multipart form: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Fatalf("Failed to copy file into request: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"?contRep="+testBucket+"&docId="+docID, &body)
	if err != nil {
		t.Fatalf("Upload request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Upload returned status %d", resp.StatusCode)
	}

	// ---- Download ----
	resp, err = client.Get(baseURL + "?get&contRep=" + testBucket + "&docId=" + docID)
	if err != nil {
		t.Fatalf("Download request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Download returned status %d", resp.StatusCode)
	}

	downloaded, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Reading downloaded data failed: %v", err)
	}

	expectedContent := `"TEST FILE TO TEST"`
	downloadedStr := strings.ReplaceAll(string(downloaded), "\r\n", "\n")
	expectedStr := strings.ReplaceAll(expectedContent, "\r\n", "\n")

	if downloadedStr != expectedStr {
		t.Fatalf("File content mismatch: expected %q got %q", expectedContent, string(downloaded))
	}

	// ---- Delete ----
	req, err = http.NewRequest("DELETE", baseURL+"?contRep="+testBucket+"&docId="+docID, nil)
	if err != nil {
		t.Fatalf("Delete request creation failed: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Delete returned status %d", resp.StatusCode)
	}
}

func TestUploadDownloadDeleteMultipart(t *testing.T) {
	//defer cleanupTestBucket() // optional extra cleanup
	docID := "FILE_WITH_SIZE_MORE_5_MB"

	// ---- Upload ----
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filePath := "FILE_WITH_SIZE_MORE_5_MB.txt"
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		t.Fatalf("Failed to create multipart form: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		t.Fatalf("Failed to copy file into request: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"?contRep="+testBucket+"&docId="+docID, &body)
	if err != nil {
		t.Fatalf("Upload request creation failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Upload returned status %d", resp.StatusCode)
	}

	// ---- Download ----
	resp, err = client.Get(baseURL + "?get&contRep=" + testBucket + "&docId=" + docID)
	if err != nil {
		t.Fatalf("Download request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Download returned status %d", resp.StatusCode)
	}

	downloaded, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Reading downloaded data failed: %v", err)
	}

	if len(downloaded) == 0 {
		t.Fatalf("File content mismatch empty result")
	}

	// ---- Delete ----
	req, err = http.NewRequest("DELETE", baseURL+"?contRep="+testBucket+"&docId="+docID, nil)
	if err != nil {
		t.Fatalf("Delete request creation failed: %v", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Delete returned status %d", resp.StatusCode)
	}
}
