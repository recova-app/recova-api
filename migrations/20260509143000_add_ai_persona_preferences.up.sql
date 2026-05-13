CREATE TABLE IF NOT EXISTS user_ai_persona_preferences (
  user_id uuid PRIMARY KEY,
  persona text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_user_ai_persona_preferences_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT chk_user_ai_persona_preferences_persona
    CHECK (persona IN ('supportive', 'friendly', 'concise', 'direct'))
);
