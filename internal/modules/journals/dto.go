package journals

// CreateJournalRequest is request payload to create one journal entry.
type CreateJournalRequest struct {
	Content string `json:"content"`
}

// CreateJournalInput is normalized journal create payload.
type CreateJournalInput struct {
	Content string
}

// ListJournalsQuery captures list-journals query parameters.
type ListJournalsQuery struct {
	Limit  *int    `query:"limit"`
	Cursor *string `query:"cursor"`
}

// JournalPayload is API payload for one journal entry.
type JournalPayload struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
