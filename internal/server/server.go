// Package server provides the HTTP handlers and routing for the MCP server.
package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"sam-mcp/internal/sam"
)

// Config contains server configuration values such as port, auth token, and API keys.
type Config struct {
	Port          string
	Token         string
	SamAPIKey     string
	ScheduleToken string
	PrefetchQ     string
	PrefetchNAICS []string
	PrefetchDays  int
	PrefetchLimit int
	PrefetchType  string
	PrefetchOrg   string
}

// Server contains the configured router, cache, HTTP client, and config for the MCP server.
type Server struct {
	cfg         Config
	router      *chi.Mux
	cache       *Cache
	httpClient  *http.Client
	toolHandlers map[string]http.HandlerFunc
}

// New constructs a Server with middleware and routes configured.
func New(cfg Config) *Server {
	s := &Server{
		cfg:        cfg,
		router:     chi.NewRouter(),
		cache:      NewCache(),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Get("/health", s.handleHealth)

	s.router.Route("/mcp", func(r chi.Router) {
		r.Use(s.auth)
		r.Get("/tools", s.handleListTools)
		r.Post("/call", s.handleCall)
		r.Post("/scheduled", s.handleScheduled)
	})

	s.registerToolHandlers()

	return s
}

func (s *Server) registerToolHandlers() {
	s.toolHandlers = map[string]http.HandlerFunc{
		"sam_search": s.handleSamSearch,
	}
}

// Router exposes the root HTTP handler for the server.
func (s *Server) Router() http.Handler { return s.router }

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.Token == "" {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+s.cfg.Token {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Tool describes an MCP tool and its input schema.
// Note: Tool and CallRequest types are defined in types.go

func (s *Server) handleListTools(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tools := []Tool{
		{
			Name:        "sam_search",
			Description: "Search SAM.gov opportunities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"q":            map[string]interface{}{"type": "string"},
					"naics":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"days":         map[string]interface{}{"type": "integer", "minimum": 0},
					"limit":        map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 100},
					"noticeType":   map[string]interface{}{"type": "string"},
					"organization": map[string]interface{}{"type": "string"},
				},
				"required": []string{"days"},
			},
		},
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tools": tools})
}

func (s *Server) handleCall(w http.ResponseWriter, r *http.Request) {
	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if handler, ok := s.toolHandlers[req.Name]; ok {
		// We need to marshal the args back into the request body
		// so the handler can decode it.
		jsonArgs, err := json.Marshal(req.Args)
		if err != nil {
			http.Error(w, "invalid arguments", http.StatusBadRequest)
			return
		}
		r.Body = http.NoBody
		newReq := r.WithContext(r.Context())
		newReq.Body = io.NopCloser(bytes.NewReader(jsonArgs))
		handler.ServeHTTP(w, newReq)
		return
	}

	http.Error(w, "unknown tool", http.StatusNotFound)
}

func (s *Server) handleSamSearch(w http.ResponseWriter, r *http.Request) {
	// If API key is configured, call SAM.gov; else return mock data
	type args struct {
		Q          string   `json:"q"`
		NAICS      []string `json:"naics"`
		Days       int      `json:"days"`
		Limit      int      `json:"limit"`
		NoticeType string   `json:"noticeType"`
		Org        string   `json:"organization"`
	}
	var a args
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	cacheKey := "sam_search:" + a.Q + ":" + time.Now().UTC().Format("2006-01-02")
	if v, ok := s.cache.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
		return
	}
	if s.cfg.SamAPIKey != "" {
		client := sam.New("https://api.sam.gov/opportunities/v2/search", s.cfg.SamAPIKey, s.httpClient)
		res, err := client.Search(r.Context(), sam.SearchParams{
			Q: a.Q, NAICS: a.NAICS, Days: a.Days, Limit: a.Limit, NoticeType: a.NoticeType, Org: a.Org,
		})
		if err != nil {
			http.Error(w, "sam api error: "+err.Error(), http.StatusBadGateway)
			return
		}
		resp := map[string]interface{}{"results": res}
		s.cache.Set(cacheKey, resp, 12*time.Hour)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	// Fallback mock
	resp := map[string]interface{}{
		"results": []map[string]string{
			{"title": "Example Opportunity", "agency": "GSA", "modified": time.Now().UTC().Format(time.RFC3339), "url": "https://sam.gov/opp/example"},
		},
	}
	s.cache.Set(cacheKey, resp, 12*time.Hour)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// handleScheduled is intended to be called by a scheduler (e.g., GitHub Actions) to warm caches or trigger background work
func (s *Server) handleScheduled(w http.ResponseWriter, _ *http.Request) {
	// For now, just refresh the mock cache entry
	resp := map[string]interface{}{
		"results": []map[string]string{
			{"title": "Example Opportunity", "agency": "GSA", "modified": time.Now().UTC().Format(time.RFC3339), "url": "https://sam.gov/opp/example"},
		},
	}
	s.cache.Set("sam_search_mock", resp, 12*time.Hour)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "scheduled task completed"})
}
