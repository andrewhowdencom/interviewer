package interview

// Config represents the top-level configuration structure.
type Config struct {
	Interviews []Topic `mapstructure:"interviews"`
}

// Topic defines the structure for a single interview topic.
type Topic struct {
	ID        string   `mapstructure:"id"`
	Name      string   `mapstructure:"name"`
	Provider  string   `mapstructure:"provider"`
	Questions []string `mapstructure:"questions,omitempty"`
	Prompt    string   `mapstructure:"prompt,omitempty"`
}
