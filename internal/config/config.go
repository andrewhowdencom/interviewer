package config

// Config defines the structure of the application's configuration file.
type Config struct {
	// DNSServer specifies a custom DNS server to use for all outbound connections.
	// If empty, the system's default DNS resolver will be used.
	DNSServer string `mapstructure:"dns_server,omitempty"`
	Telemetry struct {
		OTLP struct {
			Endpoint string
			Insecure bool
			Headers  map[string]string
		}
	}
	Interviews []Topic
	Providers  struct {
		Gemini struct {
			APIKey      string `mapstructure:"api_key"`
			Model       string
			Interviewer struct {
				Prompt string
			}
		}
	}
}

// Topic defines the structure of an interview topic.
type Topic struct {
	ID        string
	Name      string
	Provider  string
	Prompt    string
	Questions []string
}
