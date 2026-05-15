DROP INDEX IF EXISTS uq_daily_physical_challenges_title_description;
DROP TABLE IF EXISTS daily_physical_challenges;

ALTER TABLE daily_challenges
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS title;
