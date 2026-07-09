package llm

import (
	"context"
	"encoding/json"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

type Message struct {
	Role       Role
	Content    string
	ToolCallID string
	ToolCalls  []ToolCall
}

type ToolSpec struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

type ChatRequest struct {
	Message []Message
	Tools   []ToolSpec
}

type ChatResponse struct {
	Message Message
}

type LLM interface {
	Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error)
}
