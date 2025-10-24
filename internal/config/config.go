package config

// Config defines the structure of the application's configuration file.
type Config struct {
	Interviews []Topic
	Providers  struct {
		Gemini struct {
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
