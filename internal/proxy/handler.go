package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewHandler(upstreamUrl string) http.Handler {
	parsedUrl, err := url.Parse(upstreamUrl)
	if err != nil {
		panic("Failed to parse URL: " + err.Error())
	}

	return httputil.NewSingleHostReverseProxy(parsedUrl)
}
