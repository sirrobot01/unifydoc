// Package server implements the unifidoc development server with live reload.
package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Server is a static file server for generated documentation with optional
// file-watching and browser live reload.
type Server struct {
	// Dir is the output directory to serve (e.g. ./docs).
	Dir string
	// Port is the TCP port to listen on.
	Port int
	// LiveReload enables spec watching and browser auto-reload.
	LiveReload bool
	// Watch is the set of files/directories to watch for changes.
	Watch []string
	// Rebuild regenerates the documentation. It is called on each detected
	// change before clients are told to reload.
	Rebuild func() error
	// Quiet suppresses informational logging.
	Quiet bool

	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

// liveReloadScript is injected before </body> in served HTML pages. It opens an
// SSE connection and reloads the page when the server signals a rebuild.
const liveReloadScript = `
<script>
(function () {
  var es = new EventSource("/__livereload");
  es.onmessage = function () { window.location.reload(); };
  es.onerror = function () { es.close(); setTimeout(function () { window.location.reload(); }, 2000); };
})();
</script>
`

// ListenAndServe builds the initial docs, starts the watcher (if enabled), and
// blocks serving HTTP until the process is interrupted.
func (s *Server) ListenAndServe() error {
	s.clients = make(map[chan struct{}]struct{})

	if s.LiveReload {
		if err := s.startWatcher(); err != nil {
			return fmt.Errorf("failed to start watcher: %w", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/__livereload", s.handleLiveReload)
	mux.HandleFunc("/", s.handleFile)

	addr := fmt.Sprintf(":%d", s.Port)
	if !s.Quiet {
		fmt.Printf("Serving %s at http://localhost:%d\n", s.Dir, s.Port)
		if s.LiveReload {
			fmt.Println("Live reload enabled — watching for spec changes...")
		}
	}

	return http.ListenAndServe(addr, mux)
}

// handleFile serves a file from the output directory, injecting the live-reload
// script into HTML responses when enabled.
func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	// Resolve the request path safely inside Dir (path traversal protection).
	clean := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
	target := filepath.Join(s.Dir, clean)

	info, err := os.Stat(target)
	if err == nil && info.IsDir() {
		target = filepath.Join(target, "index.html")
		info, err = os.Stat(target)
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Non-HTML files are served as-is.
	if !strings.HasSuffix(strings.ToLower(target), ".html") {
		http.ServeFile(w, r, target)
		return
	}

	data, err := os.ReadFile(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if s.LiveReload {
		data = injectLiveReload(data)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}

// injectLiveReload inserts the reload script before the closing </body> tag,
// falling back to appending it if no such tag exists.
func injectLiveReload(html []byte) []byte {
	marker := []byte("</body>")
	if idx := bytes.LastIndex(html, marker); idx != -1 {
		out := make([]byte, 0, len(html)+len(liveReloadScript))
		out = append(out, html[:idx]...)
		out = append(out, []byte(liveReloadScript)...)
		out = append(out, html[idx:]...)
		return out
	}
	return append(html, []byte(liveReloadScript)...)
}

// handleLiveReload streams reload events to a connected browser via SSE.
func (s *Server) handleLiveReload(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			_, _ = fmt.Fprint(w, "data: reload\n\n")
			flusher.Flush()
		}
	}
}

// broadcast signals all connected browsers to reload.
func (s *Server) broadcast() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// startWatcher watches the configured spec files and rebuilds on change,
// debouncing rapid successive events.
func (s *Server) startWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Watch each target's parent directory so renames/atomic saves are caught.
	watched := make(map[string]struct{})
	for _, path := range s.Watch {
		dir := filepath.Dir(path)
		if _, ok := watched[dir]; ok {
			continue
		}
		if err := watcher.Add(dir); err != nil {
			if !s.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: cannot watch %s: %v\n", dir, err)
			}
			continue
		}
		watched[dir] = struct{}{}
	}

	go func() {
		var timer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
					continue
				}
				// Debounce: coalesce bursts of events into a single rebuild.
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(200*time.Millisecond, s.rebuildAndNotify)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if !s.Quiet {
					fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
				}
			}
		}
	}()

	return nil
}

// rebuildAndNotify regenerates the docs and, on success, tells browsers to reload.
func (s *Server) rebuildAndNotify() {
	if !s.Quiet {
		fmt.Println("Change detected, regenerating...")
	}
	if err := s.Rebuild(); err != nil {
		fmt.Fprintf(os.Stderr, "rebuild failed: %v\n", err)
		return
	}
	s.broadcast()
	if !s.Quiet {
		fmt.Println("✓ Reloaded")
	}
}
