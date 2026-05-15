ALTER TABLE users
  DROP CONSTRAINT IF EXISTS chk_users_porn_free_goal_positive;

ALTER TABLE users
  DROP COLUMN IF EXISTS porn_free_goal;
