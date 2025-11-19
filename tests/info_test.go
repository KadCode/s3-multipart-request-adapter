package tests

import (
	"net/http"
	"testing"
)

// TestInfoNotFound verifies that requesting non-existent doc returns 404
func TestInfoNotFound(t *testing.T) {
	resp, err := client.Get(baseURL + "?info&contRep=" + testBucket + "&docId=NO_SUCH_FILE_123")
	if err != nil {
		t.Fatalf("Info request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404 Not Found, got %d", resp.StatusCode)
	}
}
