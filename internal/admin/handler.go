package admin

import "net/http"

func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("Admin interface is under construction"))
		if err != nil {
			return
		}
	})
}
