package interview

// Config represents the top-level configuration structure.
type Config struct {
	Providers  Providers `mapstructure:"providers"`
	Interviews []Topic   `mapstructure:"interviews"`
}

// Providers defines the structure for provider configurations.
type Providers struct {
	Gemini Gemini `mapstructure:"gemini"`
}

// Gemini defines the structure for Gemini provider configuration.
type Gemini struct {
	Model string `mapstructure:"model"`
}

// Topic defines the structure for a single interview topic.
type Topic struct {
	ID        string   `mapstructure:"id"`
	Name      string   `mapstructure:"name"`
	Provider  string   `mapstructure:"provider"`
	Questions []string `mapstructure:"questions,omitempty"`
	Prompt    string   `mapstructure:"prompt,omitempty"`
}
