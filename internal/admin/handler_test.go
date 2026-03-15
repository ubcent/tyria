package admin

import (
	"io"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	ts := httptest.NewServer(NewHandler())
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL)

	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Admin interface is under construction" {
		t.Fatalf("Expected response body 'Admin interface is under construction', got '%s'", string(body))
	}
}
