build:
	go build -o bin/sam-mcp ./cmd/sam-mcp-http

run:
	PORT=3000 MCP_TOKEN=devtoken go run ./cmd/sam-mcp-http

test:
	go test ./...

docker:
	docker build -t sam-mcp:local .

compose:
	docker compose up --build
