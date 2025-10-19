package server

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHealth(t *testing.T) {
    s := New(Config{})
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rr := httptest.NewRecorder()
    s.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
}

func TestToolsAndCall(t *testing.T) {
    s := New(Config{Token: "x"})

    // Unauthorized
    req := httptest.NewRequest(http.MethodGet, "/mcp/tools", nil)
    rr := httptest.NewRecorder()
    s.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rr.Code)
    }

    // Authorized tools
    req = httptest.NewRequest(http.MethodGet, "/mcp/tools", nil)
    req.Header.Set("Authorization", "Bearer x")
    rr = httptest.NewRecorder()
    s.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }

    // Call sam_search
    body, _ := json.Marshal(map[string]interface{}{"name": "sam_search", "arguments": map[string]interface{}{"days": 7}})
    req = httptest.NewRequest(http.MethodPost, "/mcp/call", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer x")
    rr = httptest.NewRecorder()
    s.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
}

func TestScheduled(t *testing.T) {
    s := New(Config{Token: "x"})
    req := httptest.NewRequest(http.MethodPost, "/mcp/scheduled", nil)
    req.Header.Set("Authorization", "Bearer x")
    rr := httptest.NewRecorder()
    s.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }
}

func TestSamSearchMock(t *testing.T) {
    s := New(Config{})
    body, _ := json.Marshal(map[string]interface{}{"days": 7})
    req := httptest.NewRequest(http.MethodPost, "/mcp/call", bytes.NewReader(body))
    rr := httptest.NewRecorder()
    s.handleSamSearch(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rr.Code)
    }

    var resp map[string]interface{}
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("invalid json: %v", err)
    }
    if _, ok := resp["results"]; !ok {
        t.Fatal("expected results key in response")
    }
}
