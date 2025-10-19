#!/usr/bin/env bash
set -euo pipefail
export PORT=${PORT:-3000}
export MCP_TOKEN=${MCP_TOKEN:-devtoken}
export ALLOW_INSECURE_HTTP=1
exec go run ./cmd/sam-mcp-http
