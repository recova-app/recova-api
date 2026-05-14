CREATE TABLE IF NOT EXISTS relapses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL,
  check_in_id uuid NULL,
  relapse_date date NOT NULL,
  mood varchar(50) NOT NULL,
  commitment text NULL,
  relapse_trigger text[] NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_relapses_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON DELETE CASCADE,
  CONSTRAINT fk_relapses_check_in
    FOREIGN KEY (check_in_id) REFERENCES check_ins(id)
    ON DELETE SET NULL,
  CONSTRAINT uq_relapses_user_date UNIQUE(user_id, relapse_date)
);

CREATE INDEX IF NOT EXISTS idx_relapses_user_id ON relapses(user_id);
CREATE INDEX IF NOT EXISTS idx_relapses_date ON relapses(relapse_date);
