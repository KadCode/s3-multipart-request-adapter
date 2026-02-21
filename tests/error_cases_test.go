package tests

import (
	"io"
	"net/http"
	"testing"
)

// TestMissingDocID ensures request fails when docId is missing
func TestMissingDocID(t *testing.T) {
	resp, err := client.Get(baseURL + "?get&contRep=" + testBucket)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 400, got %d resp body %s", resp.StatusCode, string(bodyBytes))
	}
}

// TestUnknownAction ensures unknown actions return 400
func TestUnknownAction(t *testing.T) {
	resp, err := client.Get(baseURL + "?contRep=" + testBucket)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", resp.StatusCode)
	}
}
