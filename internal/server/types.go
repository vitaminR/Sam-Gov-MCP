package server

type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

type CallRequest struct {
    Name   string                 `json:"name"`
    Args   map[string]interface{} `json:"arguments"`
}
