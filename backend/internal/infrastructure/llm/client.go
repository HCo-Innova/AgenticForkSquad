package llm

type LLMClient interface {
	SendMessage(prompt, system string) (string, error)
	SendMessageWithJSON(prompt, system string) (map[string]interface{}, error)
	GetUsage() (inputTokens, outputTokens int)
}
