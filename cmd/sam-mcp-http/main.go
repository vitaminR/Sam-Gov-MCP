// Command sam-mcp-http starts the MCP HTTP server.
package main

import (
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    "sam-mcp/internal/server"
)

func main() {
    cfg := server.Config{
        Port: getEnv("PORT", "3000"),
        Token: os.Getenv("MCP_TOKEN"),
        SamAPIKey: os.Getenv("SAM_API_KEY"),
        ScheduleToken: os.Getenv("SCHEDULE_TOKEN"),
        PrefetchQ: os.Getenv("PREFETCH_Q"),
        PrefetchNAICS: splitCSV(os.Getenv("PREFETCH_NAICS")),
        PrefetchDays: getEnvInt("PREFETCH_DAYS", 7),
        PrefetchLimit: getEnvInt("PREFETCH_LIMIT", 25),
        PrefetchType: os.Getenv("PREFETCH_NOTICE_TYPE"),
        PrefetchOrg: os.Getenv("PREFETCH_ORG"),
    }
    if cfg.Token == "" {
        log.Println("WARN: MCP_TOKEN not set; endpoints will be open. Set MCP_TOKEN to secure.")
    }
    if cfg.SamAPIKey == "" {
        log.Println("INFO: SAM_API_KEY not set; sam_search will use mock data until configured.")
    }
    srv := server.New(cfg)
    log.Printf("Starting MCP HTTP server on :%s\n", cfg.Port)
    certFile := os.Getenv("TLS_CERT_FILE")
    keyFile := os.Getenv("TLS_KEY_FILE")
    if certFile == "" || keyFile == "" {
        log.Fatal("TLS_CERT_FILE and TLS_KEY_FILE are required. Provide TLS cert/key or run behind a TLS-terminating proxy.")
    }
    log.Println("TLS enabled: using provided certificate and key")
    if err := http.ListenAndServeTLS(":"+cfg.Port, certFile, keyFile, srv.Router()); err != nil {
        log.Fatalf("server error: %v", err)
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func getEnvInt(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        if i, err := strconv.Atoi(v); err == nil {
            return i
        }
    }
    return def
}

func splitCSV(v string) []string {
    if v == "" { return nil }
    parts := strings.Split(v, ",")
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p != "" { out = append(out, p) }
    }
    return out
}
