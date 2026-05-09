CREATE TABLE IF NOT EXISTS achievements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  title text NOT NULL,
  description text NOT NULL,
  category text NOT NULL,
  threshold numeric NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_achievements_threshold_non_negative CHECK (threshold >= 0)
);

CREATE TABLE IF NOT EXISTS user_achievement_progress (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL,
  achievement_id uuid NOT NULL,
  progress_value numeric NOT NULL DEFAULT 0,
  unlocked_at timestamptz,
  last_evaluated_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_user_achievement_progress_user
    FOREIGN KEY (user_id) REFERENCES users(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT fk_user_achievement_progress_achievement
    FOREIGN KEY (achievement_id) REFERENCES achievements(id)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT uq_user_achievement_progress_user_achievement
    UNIQUE (user_id, achievement_id),
  CONSTRAINT chk_user_achievement_progress_non_negative
    CHECK (progress_value >= 0)
);

CREATE INDEX IF NOT EXISTS idx_user_achievement_progress_user
  ON user_achievement_progress(user_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_achievement_progress_achievement
  ON user_achievement_progress(achievement_id);

CREATE INDEX IF NOT EXISTS idx_user_achievement_progress_user_unlocked
  ON user_achievement_progress(user_id, unlocked_at DESC)
  WHERE unlocked_at IS NOT NULL;
