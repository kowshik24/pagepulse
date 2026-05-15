package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"pagepulse/internal/buildinfo"
	"pagepulse/internal/metrics"
)

//go:embed static/*
var embeddedWeb embed.FS

type Server struct {
	collector *metrics.Collector
	buildInfo buildinfo.Info
	mux       *http.ServeMux
	clientsMu sync.Mutex
	clients   map[chan metrics.Summary]struct{}
}

func NewServer(collector *metrics.Collector, info buildinfo.Info) (*Server, error) {
	assetFS, err := fs.Sub(embeddedWeb, "static")
	if err != nil {
		return nil, err
	}

	s := &Server{
		collector: collector,
		buildInfo: info,
		mux:       http.NewServeMux(),
		clients:   map[chan metrics.Summary]struct{}{},
	}

	s.mux.Handle("/", http.FileServer(http.FS(assetFS)))
	s.mux.HandleFunc("/api/v1/summary", s.handleSummary)
	s.mux.HandleFunc("/api/v1/resources", s.handleResources)
	s.mux.HandleFunc("/api/v1/version", s.handleVersion)
	s.mux.HandleFunc("/api/v1/stream", s.handleStream)

	go s.broadcastLoop()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return withDefaultHeaders(s.mux)
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.collector.Summary())
}

func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.collector.Resources())
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.buildInfo)
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan metrics.Summary, 4)
	s.addClient(ch)
	defer s.removeClient(ch)

	s.writeSSE(w, s.collector.Summary())
	flusher.Flush()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ssum := <-ch:
			s.writeSSE(w, ssum)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) writeSSE(w http.ResponseWriter, payload metrics.Summary) {
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: summary\ndata: %s\n\n", b)
}

func (s *Server) broadcastLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		sum := s.collector.Summary()
		s.clientsMu.Lock()
		for ch := range s.clients {
			select {
			case ch <- sum:
			default:
			}
		}
		s.clientsMu.Unlock()
	}
}

func (s *Server) addClient(ch chan metrics.Summary) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[ch] = struct{}{}
}

func (s *Server) removeClient(ch chan metrics.Summary) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	delete(s.clients, ch)
	close(ch)
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func withDefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func RunHTTP(ctx context.Context, server *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}
