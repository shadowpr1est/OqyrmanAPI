package llm

import "context"

type Message struct {
	Role    string
	Content string
}

// StreamCallback вызывается для каждого фрагмента текста при стриминге.
// Если возвращает ошибку — стрим прерывается.
type StreamCallback func(chunk string) error

type LLMClient interface {
	Complete(ctx context.Context, system string, messages []Message) (string, error)
	CompleteStream(ctx context.Context, system string, messages []Message, cb StreamCallback) error
}
