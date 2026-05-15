ALTER TABLE users
  ADD COLUMN IF NOT EXISTS porn_free_goal integer;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_users_porn_free_goal_positive'
  ) THEN
    ALTER TABLE users
      ADD CONSTRAINT chk_users_porn_free_goal_positive
      CHECK (porn_free_goal IS NULL OR porn_free_goal > 0);
  END IF;
END $$;
