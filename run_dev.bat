@echo off
set PORT=%PORT: =3000%
if "%MCP_TOKEN%"=="" set MCP_TOKEN=devtoken
set ALLOW_INSECURE_HTTP=1
go run ./cmd/sam-mcp-http
