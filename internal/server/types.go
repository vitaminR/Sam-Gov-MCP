package server

// Tool describes an MCP tool and its input schema exposed by this server.
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// CallRequest represents a request to invoke a specific MCP tool with arguments.
type CallRequest struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"arguments"`
}
