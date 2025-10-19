# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/sam-mcp ./cmd/sam-mcp-http

FROM gcr.io/distroless/static:nonroot
WORKDIR /
ENV PORT=3000
COPY --from=build /out/sam-mcp /sam-mcp
USER nonroot:nonroot
EXPOSE 3000
ENTRYPOINT ["/sam-mcp"]
