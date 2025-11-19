package tests

import (
	"net/http"
	"testing"
)

// TestList verifies that listing objects returns 200 OK
func TestList(t *testing.T) {
	resp, err := client.Get(baseURL + "?list&contRep=" + testBucket)
	if err != nil {
		t.Fatalf("List request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
