package ctxkeys

type contextKey string

const (
	UserAgentKey contextKey = "user_agent"
	ClientIPKey  contextKey = "client_ip"
)
