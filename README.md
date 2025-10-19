SAM MCP Server (Go)

Overview
- Minimal Model Context Protocol (MCP) HTTP server in Go
- Endpoints: /health, /mcp/tools, /mcp/call, /mcp/scheduled
- Token auth on /mcp/* via Authorization: Bearer <MCP_TOKEN>
- Docker, unit tests, CI, and twice-daily schedule

Run locally
- Env: PORT=3000 (default), MCP_TOKEN=your-secret
- go run ./cmd/sam-mcp-http

Docker
- docker compose up --build

Endpoints
- GET /health -> {"status":"ok"}
- GET /mcp/tools (auth required)
- POST /mcp/call (body: {"name":"sam_search","arguments":{"days":7}})
- POST /mcp/scheduled (for cache warmup)

Agent Builder
- URL: https://<host>/mcp
- Header: Authorization: Bearer ${env:MCP_TOKEN}

Scheduling (twice a day)
- .github/workflows/scheduled.yml uses cron at 06:00 and 18:00 UTC
- After you deploy, set secret MCP_SCHEDULE_URL and enable the curl step to hit /mcp/scheduled

Tests
- go test ./...

Notes
- sam_search returns a mock result with a 12h cache; replace with real SAM.gov integration later.