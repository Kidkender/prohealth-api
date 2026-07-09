package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/sony/gobreaker/v2"
)

type OpenAIProvider struct {
	client  openai.Client
	model   string
	breaker *gobreaker.CircuitBreaker[*ChatResponse]
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	breaker := gobreaker.NewCircuitBreaker[*ChatResponse](gobreaker.Settings{
		Name:        "openai",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.Warn("circuit breaker change", "name", name, "from", from, "to", to)
		},
	})
	return &OpenAIProvider{
		client:  openai.NewClient(option.WithAPIKey(apiKey)),
		model:   model,
		breaker: breaker,
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return p.breaker.Execute(func() (*ChatResponse, error) {
		return p.DoChat(ctx, req)
	})
}

func toResponseInput(msgs []Message) responses.ResponseInputParam {
	items := make(responses.ResponseInputParam, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleTool:
			items = append(items, responses.ResponseInputItemParamOfFunctionCallOutput(m.ToolCallID, m.Content))
		case RoleAssistant:
			if len(m.ToolCalls) > 0 {
				for _, tc := range m.ToolCalls {
					items = append(items, responses.ResponseInputItemParamOfFunctionCall(string(tc.Args), tc.ID, tc.Name))
				}
				continue
			}
			items = append(items, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleAssistant))
		case RoleSystem:
			items = append(items, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleSystem))
		default: // RoleUser
			items = append(items, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleUser))
		}
	}
	return items
}

func toResponseTools(specs []ToolSpec) ([]responses.ToolUnionParam, error) {
	tools := make([]responses.ToolUnionParam, 0, len(specs))
	for _, spec := range specs {
		var params map[string]any
		if len(spec.Parameters) > 0 {
			if err := json.Unmarshal(spec.Parameters, &params); err != nil {
				return nil, fmt.Errorf("invalid parameters schema for tool %q: %w", spec.Name, err)
			}
		}
		tools = append(tools, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        spec.Name,
				Description: openai.String(spec.Description),
				Parameters:  params,
				Strict:      openai.Bool(false),
			},
		})
	}
	return tools, nil
}

func toChatResponse(resp *responses.Response) *ChatResponse {
	var toolCalls []ToolCall
	for _, item := range resp.Output {
		if item.Type != "function_call" {
			continue
		}
		toolCalls = append(toolCalls, ToolCall{
			ID:   item.CallID,
			Name: item.Name,
			Args: json.RawMessage(item.Arguments),
		})
	}

	return &ChatResponse{
		Message: Message{
			Role:      RoleAssistant,
			Content:   resp.OutputText(),
			ToolCalls: toolCalls,
		},
	}
}

func (p *OpenAIProvider) DoChat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	tools, err := toResponseTools(req.Tools)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: toResponseInput(req.Message)},
		Model: openai.ChatModelGPT4o,
		Tools: tools,
	})
	if err != nil {
		return nil, err
	}
	return toChatResponse(resp), nil
}
