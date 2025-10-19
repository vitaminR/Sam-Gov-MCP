SAM.gov MCP Server (Go)

Overview

- Production-ready HTTP MCP server implementing the Model Context Protocol
- Integrates with SAM.gov Opportunities API with a 12h in-memory cache
- Endpoints: /health, /mcp/tools, /mcp/call, /mcp/scheduled
- Bearer token auth on /mcp/\*; separate schedule token for /mcp/scheduled
- Docker container, GitHub Actions CI, and twice-daily scheduler

Architecture

- cmd/sam-mcp-http: main entrypoint, reads env, wires server and TLS
- internal/server:
  - server.go: routing, auth middleware, handlers (tools, call, scheduled)
  - types.go: Tool, CallRequest shapes for MCP
  - cache.go: simple thread-safe TTL cache
  - sam.go: minimal HTTP client for SAM.gov opportunities search
- internal/sam: richer SAM.gov client used by server handler

Security

- MCP_TOKEN protects /mcp/tools and /mcp/call
- SCHEDULE_TOKEN protects /mcp/scheduled (may also accept MCP_TOKEN)
- TLS is recommended for all deployments; compose mounts certificates

Environment variables

- PORT: server port (default 3000)
- MCP_TOKEN: bearer token for MCP endpoints
- SCHEDULE_TOKEN: bearer token for scheduled endpoint
- SAM_API_KEY: API key for SAM.gov (optional; if unset, mock data is returned)
- PREFETCH_Q: default query for scheduled prefetch (e.g., "software")
- PREFETCH_NAICS: CSV NAICS codes (e.g., 541511,541512,541519)
- PREFETCH_DAYS: integer days back to search (e.g., 7)
- PREFETCH_LIMIT: integer page size (e.g., 25)
- PREFETCH_NOTICE_TYPE: optional notice type filter
- PREFETCH_ORG: optional organization filter
- TLS_CERT_FILE: path to server certificate (PEM)
- TLS_KEY_FILE: path to server key (PEM)

Run locally (Go)

1. Copy .env.local and edit values (at least MCP_TOKEN and SCHEDULE_TOKEN)
2. Option A: HTTP (dev only)

- go run ./cmd/sam-mcp-http

3. Option B: HTTPS (recommended)

- Generate certs into ./certs (see TLS below) and set TLS_CERT_FILE/TLS_KEY_FILE
- go run ./cmd/sam-mcp-http

Docker

- docker compose up --build
- Compose loads .env.local, mounts ./certs, and passes environment to the container

TLS (local self-signed)
Generate development certs into ./certs:

- certs/server.crt
- certs/server.key
  Then set:
- TLS_CERT_FILE=./certs/server.crt
- TLS_KEY_FILE=./certs/server.key

HTTP endpoints

- GET /health
  - 200 {"status":"ok"}
- GET /mcp/tools (auth: Authorization: Bearer <MCP_TOKEN>)
  - Lists available tools and input schemas
- POST /mcp/call (auth)
  - Body: {"name":"sam_search","arguments":{...}}
  - Routes request to tool handler
- POST /mcp/scheduled (auth: Bearer <SCHEDULE_TOKEN> or MCP_TOKEN)
  - Triggers cache warm-up using PREFETCH\_\* defaults

Tool: sam_search
Input arguments (all optional unless specified):

- q: string (search text)
- naics: string[]
- days: integer (required by default schema)
- limit: integer (1..100)
- noticeType: string
- organization: string

Curl examples
List tools:
curl -H "Authorization: Bearer $MCP_TOKEN" https://<host>/mcp/tools

Call tool:
curl -H "Authorization: Bearer $MCP_TOKEN" \
 -H "Content-Type: application/json" \
 -d '{"name":"sam_search","arguments":{"q":"software","days":7,"limit":25}}' \
 https://<host>/mcp/call

Trigger scheduled prefetch:
curl -H "Authorization: Bearer $SCHEDULE_TOKEN" https://<host>/mcp/scheduled

GitHub Actions scheduler

- .github/workflows/scheduled.yml runs at 06:00 and 18:00 UTC
- Set these repository secrets after deployment:
  - MCP_SCHEDULE_URL: full URL to POST (e.g., https://<host>/mcp/scheduled)
  - SCHEDULE_TOKEN: bearer token for scheduled auth
- The step is conditional and only runs when both secrets are set

CI

- On push/PR, the CI workflow builds and vets the code
- Trivy, revive, and Semgrep are used via Codacy MCP tooling

Dev quickstart (no TLS, for tunneling)

- For fast Agent Builder testing via an HTTPS tunnel:
  - Linux/WSL: ./run_dev.sh
  - Windows: run_dev.bat
  - This sets ALLOW_INSECURE_HTTP=1 and starts on <http://localhost:3000>
  - Then expose with your HTTPS tunnel/proxy to https://<host>/mcp

Testing

- go test ./...

Using with OpenAI Agent Builder (MCP)
You can connect this server as an MCP tool in OpenAI Agent Builder.

Prerequisites

- Deployed HTTPS endpoint reachable by OpenAI
- MCP_TOKEN configured

Steps

1. Open the Agent Builder UI (platform.openai.com/agents) and create/edit your agent
1. Go to Tools > Add Tool > Model Context Protocol (MCP)
1. Enter:

- Base URL: https://<host>/mcp
- Auth: HTTP header
- Header name: Authorization
- Header value: Bearer ${MCP_TOKEN}

1. Save the tool and test:

- Ask the agent: "Search SAM.gov for software opportunities from the last 7 days"
- The agent will call sam_search with your arguments

Troubleshooting

- 401 Unauthorized: verify Authorization header and token values
- Empty results: ensure SAM_API_KEY is set if you expect live data; otherwise mock results are returned
- Scheduler not firing: confirm secrets MCP_SCHEDULE_URL and SCHEDULE_TOKEN are set and URL is reachable
- TLS issues: verify cert/key paths and that the certificate matches the hostname
