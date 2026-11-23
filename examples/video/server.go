//go:build !js

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Only the packaged site assets and the fixed-size composition renderer are
// exposed. A disconnected client cancels its encoder and releases the GPU.
func serveWebsite(dir string, port int) error {
	mux := http.NewServeMux()
	for _, file := range []string{"index.html", "wasm.js", "main.wasm"} {
		path, err := filepath.Abs(filepath.Join(dir, file))
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("build the web bundle first: %w", err)
		}
		route := "/" + file
		if file == "index.html" {
			route = "/{$}"
		}
		mux.HandleFunc("GET "+route, func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, path) })
	}
	rendering := make(chan struct{}, 1)
	mux.HandleFunc("POST /render", func(w http.ResponseWriter, r *http.Request) {
		c := composition{-1, -1}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512)).Decode(&c); err != nil || !c.valid() {
			http.Error(w, "Invalid composition.", http.StatusBadRequest)
			return
		}
		select {
		case rendering <- struct{}{}:
			defer func() { <-rendering }()
		default:
			http.Error(w, "Another export is running. Try again when it finishes.", http.StatusTooManyRequests)
			return
		}
		path, err := renderMovie(r.Context(), "", c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(path)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", templates[c.Scene].name+".mp4"))
		w.Header().Set("Content-Type", "video/mp4")
		http.ServeFile(w, r, path)
	})
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("Keel Motion at http://%s\n", addr)
	server := http.Server{Addr: addr, Handler: http.NewCrossOriginProtection().Handler(mux), ReadHeaderTimeout: 5 * time.Second}
	return server.ListenAndServe()
}
