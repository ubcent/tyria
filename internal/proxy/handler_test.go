package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyForwarding(t *testing.T) {
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Hello, from the upstream server!"))
		if err != nil {
			return
		}
	}))
	defer upstreamServer.Close()

	ts := httptest.NewServer(NewHandler(upstreamServer.URL))
	defer ts.Close()

	res, err := http.Get(ts.URL)

	if err != nil {
		t.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)

	if res.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "Hello, from the upstream server!" {
		t.Fatalf("Expected response body 'Hello, from the upstream server!', got '%s'", string(body))
	}
}
