// Package sam provides a minimal client for the SAM.gov Opportunities API.
package sam

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

// Client is a minimal HTTP client for SAM.gov opportunities search.
type Client struct {
    BaseURL string
    APIKey  string
    HTTP    *http.Client
}

// New returns a new client. If httpClient is nil, a default with 15s timeout is used.
func New(baseURL, apiKey string, httpClient *http.Client) *Client {
    if httpClient == nil {
        httpClient = &http.Client{Timeout: 15 * time.Second}
    }
    return &Client{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: apiKey, HTTP: httpClient}
}

// SearchParams defines supported search filters. All are optional except Days or explicit date filters.
type SearchParams struct {
    Q          string
    NAICS      []string
    Days       int
    Limit      int
    NoticeType string
    Org        string
}

// Opportunity is a small normalized view of an opportunity.
type Opportunity struct {
    Title    string    `json:"title"`
    Agency   string    `json:"agency"`
    Modified time.Time `json:"modified"`
    URL      string    `json:"url"`
    Raw      any       `json:"raw,omitempty"`
}

// Search performs a search against the opportunities API and returns normalized results.
// Note: The SAM.gov API parameters and fields may evolve; this method aims to be tolerant.
func (c *Client) Search(ctx context.Context, p SearchParams) ([]Opportunity, error) {
    if c.APIKey == "" {
        return nil, errors.New("sam api key missing")
    }
    reqURL, err := c.buildSearchURL(p)
    if err != nil { return nil, err }
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
    if err != nil { return nil, err }
    resp, err := c.HTTP.Do(req)
    if err != nil { return nil, err }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("sam api status %d", resp.StatusCode)
    }
    body, err := decodeJSON(resp)
    if err != nil { return nil, err }
    items := extractItems(body)
    return normalize(items), nil
}

func getString(m map[string]any, key string) string {
    if m == nil { return "" }
    if v, ok := m[key]; ok {
        switch t := v.(type) {
        case string:
            return t
        }
    }
    return ""
}

// buildSearchURL composes the search URL with query params.
func (c *Client) buildSearchURL(p SearchParams) (string, error) {
    u, err := url.Parse(c.BaseURL)
    if err != nil { return "", fmt.Errorf("invalid base url: %w", err) }
    q := u.Query()
    q.Set("api_key", c.APIKey)
    if p.Q != "" { q.Set("q", p.Q) }
    if len(p.NAICS) > 0 { q.Set("naics", strings.Join(p.NAICS, ",")) }
    if p.Limit > 0 { q.Set("limit", fmt.Sprintf("%d", p.Limit)) }
    if p.NoticeType != "" { q.Set("notice_type", p.NoticeType) }
    if p.Org != "" { q.Set("organization", p.Org) }
    if p.Days > 0 {
        from := time.Now().AddDate(0, 0, -p.Days).Format("2006-01-02")
        q.Set("date_modified_from", from)
        q.Set("postedFrom", from)
    }
    u.RawQuery = q.Encode()
    return u.String(), nil
}

// decodeJSON decodes an HTTP response body into a generic interface.
func decodeJSON(resp *http.Response) (any, error) {
    var body any
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { return nil, err }
    return body, nil
}

// extractItems tries common result field names or array root.
func extractItems(body any) []any {
    if m, ok := body.(map[string]any); ok {
        if v, ok := m["opportunitiesData"]; ok {
            if arr, ok := v.([]any); ok { return arr }
        }
        if v, ok := m["data"]; ok {
            if arr, ok := v.([]any); ok { return arr }
        }
        if v, ok := m["results"]; ok {
            if arr, ok := v.([]any); ok { return arr }
        }
    }
    if arr, ok := body.([]any); ok { return arr }
    return nil
}

// normalize converts raw items into Opportunities.
func normalize(items []any) []Opportunity {
    out := make([]Opportunity, 0, len(items))
    for _, it := range items {
        m, _ := it.(map[string]any)
        title := firstNonEmpty(getString(m, "title"), getString(m, "noticeTitle"))
        agency := firstNonEmpty(getString(m, "agency"), getString(m, "department"))
        urlStr := firstNonEmpty(getString(m, "uiLink"), getString(m, "url"))
        mod := parseTime(firstNonEmpty(getString(m, "lastModifiedDate"), getString(m, "dateModified")))
        out = append(out, Opportunity{Title: title, Agency: agency, Modified: mod, URL: urlStr, Raw: it})
    }
    return out
}

func firstNonEmpty(vals ...string) string {
    for _, v := range vals { if v != "" { return v } }
    return ""
}

func parseTime(s string) time.Time {
    if s == "" { return time.Time{} }
    if t, err := time.Parse(time.RFC3339, s); err == nil { return t }
    if t, err := time.Parse("2006-01-02", s); err == nil { return t }
    return time.Time{}
}
