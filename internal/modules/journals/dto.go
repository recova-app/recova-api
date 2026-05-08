package journals

// CreateJournalRequest is request payload to create one journal entry.
type CreateJournalRequest struct {
	Content string `json:"content"`
}

// CreateJournalInput is normalized journal create payload.
type CreateJournalInput struct {
	Content string
}

// JournalPayload is API payload for one journal entry.
type JournalPayload struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}
