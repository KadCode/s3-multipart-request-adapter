package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var baseURL = "https://localhost:8080/ContentServer/ContentServer.dll"
var client *http.Client

func init() {
	// Initialize TLS HTTP client
	cert, err := tls.LoadX509KeyPair("../certs/cert.pem", "../certs/key.pem")
	if err != nil {
		panic(fmt.Sprintf("failed to load cert/key: %v", err))
	}

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			},
		},
		Timeout: 10 * time.Second,
	}
}

// uploadFile uploads a file to the server using POST multipart/form-data
func uploadFile(bucket, docID, filePath string) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return fmt.Errorf("failed to copy file into request: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"?contRep="+bucket+"&docId="+docID, &body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload returned status %d", resp.StatusCode)
	}

	return nil
}

// downloadFile downloads a file and returns its contents
func downloadFile(bucket, docID string) ([]byte, error) {
	resp, err := client.Get(baseURL + "?get&contRep=" + bucket + "&docId=" + docID)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading downloaded data failed: %v", err)
	}

	return data, nil
}

// deleteFile deletes a file from the server
func deleteFile(bucket, docID string) error {
	req, err := http.NewRequest("DELETE", baseURL+"?contRep="+bucket+"&docId="+docID, nil)
	if err != nil {
		return fmt.Errorf("delete request creation failed: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete returned status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	bucket := "test-bucket"
	docID := "BENCHMARK-FILE"
	filePath := "test.txt"
	iterations := 5 // number of times to run upload/download/delete

	for i := 1; i <= iterations; i++ {
		start := time.Now()
		fmt.Printf("Iteration %d...\n", i)

		if err := uploadFile(bucket, docID, filePath); err != nil {
			fmt.Printf("Upload failed: %v\n", err)
			return
		}

		_, err := downloadFile(bucket, docID)
		if err != nil {
			fmt.Printf("Download failed: %v\n", err)
			return
		}

		if err := deleteFile(bucket, docID); err != nil {
			fmt.Printf("Delete failed: %v\n", err)
			return
		}

		fmt.Printf("Iteration %d completed in %v\n", i, time.Since(start))
	}
}
