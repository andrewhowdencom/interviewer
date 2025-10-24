package interview

// Config represents the top-level configuration structure.
type Config struct {
	Providers  Providers `mapstructure:"providers"`
	Interviews []Topic   `mapstructure:"interviews"`
}

// Providers defines the structure for provider configurations.
type Providers struct {
	Gemini Gemini `mapstructure:"gemini"`
	Slack  Slack  `mapstructure:"slack"`
}

// Slack defines the structure for Slack provider configuration.
type Slack struct {
	BotToken      string `mapstructure:"bot-token"`
	SigningSecret string `mapstructure:"signing-secret"`
}

// Gemini defines the structure for Gemini provider configuration.
type Gemini struct {
	Model       string      `mapstructure:"model"`
	Interviewer Interviewer `mapstructure:"interviewer"`
}

// Interviewer defines the structure for the interviewer configuration.
type Interviewer struct {
	Prompt string `mapstructure:"prompt"`
}

// Topic defines the structure for a single interview topic.
type Topic struct {
	ID        string   `mapstructure:"id"`
	Name      string   `mapstructure:"name"`
	Provider  string   `mapstructure:"provider"`
	Questions []string `mapstructure:"questions,omitempty"`
	Prompt    string   `mapstructure:"prompt,omitempty"`
}
