package vendors

import "testing"

func TestGroqSerializesAsOpenAICompatible(t *testing.T) {
	config := NewGroq(GroqOptions{
		APIKey:  "groq-key",
		Model:   "llama-3.3-70b-versatile",
		BaseURL: "https://api.groq.com/openai/v1/chat/completions",
	}).ToConfig()

	if config["url"] != "https://api.groq.com/openai/v1/chat/completions" {
		t.Fatalf("unexpected url: %v", config["url"])
	}
	if config["style"] != "openai" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
}

func TestCustomLLMMarksRequestAsCustom(t *testing.T) {
	config := NewCustomLLM(CustomLLMOptions{
		APIKey:  "key",
		Model:   "model",
		BaseURL: "https://llm.example.com/chat",
	}).ToConfig()

	if config["vendor"] != "custom" {
		t.Fatalf("unexpected vendor: %v", config["vendor"])
	}
	if config["style"] != "openai" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
}

func TestAnthropicSerializesRequiredClaudeFields(t *testing.T) {
	maxTokens := 1024
	config := NewAnthropic(AnthropicOptions{
		APIKey:    "anthropic-key",
		Model:     "claude-3-5-sonnet-20241022",
		URL:       "https://api.anthropic.com/v1/messages",
		Headers:   map[string]string{"anthropic-version": "2023-06-01"},
		MaxTokens: &maxTokens,
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	headers := config["headers"].(map[string]string)
	if config["url"] != "https://api.anthropic.com/v1/messages" || config["style"] != "anthropic" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if headers["anthropic-version"] != "2023-06-01" {
		t.Fatalf("unexpected headers: %#v", headers)
	}
	if params["model"] != "claude-3-5-sonnet-20241022" || params["max_tokens"] != 1024 {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestAzureOpenAIIncludesRequiredModelParam(t *testing.T) {
	config := NewAzureOpenAI(AzureOpenAIOptions{
		APIKey:         "azure-key",
		Endpoint:       "https://example.openai.azure.com",
		DeploymentName: "deployment",
		Model:          "gpt-4o",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["vendor"] != "azure" || config["style"] != "openai" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if params["model"] != "gpt-4o" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestVertexAILLMIncludesProjectRouting(t *testing.T) {
	config := NewVertexAILLM(VertexAILLMOptions{
		GeminiOptions: GeminiOptions{
			APIKey: "vertex-token",
			Model:  "gemini-2.0-flash",
		},
		ProjectID: "project",
		Location:  "us-central1",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["style"] != "gemini" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	if params["project_id"] != "project" || params["location"] != "us-central1" {
		t.Fatalf("unexpected params: %#v", params)
	}
}

func TestAmazonBedrockSerializesAsBedrockStyle(t *testing.T) {
	config := NewAmazonBedrock(AmazonBedrockOptions{
		AccessKey: "aws-access",
		SecretKey: "aws-secret",
		Region:    "us-east-1",
		Model:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
	}).ToConfig()

	if config["style"] != "bedrock" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	if config["url"] != "https://bedrock-runtime.us-east-1.amazonaws.com/model/anthropic.claude-3-5-sonnet-20241022-v2:0/converse-stream" {
		t.Fatalf("unexpected url: %v", config["url"])
	}
	if config["access_key"] != "aws-access" || config["secret_key"] != "aws-secret" || config["region"] != "us-east-1" {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestDifySerializesConversationFields(t *testing.T) {
	config := NewDify(DifyOptions{
		APIKey:         "dify-key",
		URL:            "https://api.dify.ai/v1/chat-messages",
		Model:          "default",
		User:           "user-1",
		ConversationID: "conversation-1",
	}).ToConfig()

	params := config["params"].(map[string]interface{})
	if config["style"] != "dify" {
		t.Fatalf("unexpected style: %v", config["style"])
	}
	if params["model"] != "default" {
		t.Fatalf("unexpected params: %#v", params)
	}
	if params["user"] != "user-1" || params["conversation_id"] != "conversation-1" {
		t.Fatalf("unexpected params: %#v", params)
	}
}
