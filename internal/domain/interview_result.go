package domain

import "time"

// Interview holds the metadata for an interview session.
type Interview struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Transcript holds the full question-and-answer record of an interview.
type Transcript struct {
	InterviewID string `json:"interview_id"`
	Entries     []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"entries"`
}

// Summary holds the generated summary of an interview.
type Summary struct {
	InterviewID string `json:"interview_id"`
	Text        string `json:"text"`
}
