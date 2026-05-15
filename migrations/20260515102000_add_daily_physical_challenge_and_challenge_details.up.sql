ALTER TABLE daily_challenges
  ADD COLUMN IF NOT EXISTS title text,
  ADD COLUMN IF NOT EXISTS description text;

UPDATE daily_challenges
SET
  title = COALESCE(NULLIF(trim(title), ''), 'Tantangan Harian'),
  description = COALESCE(NULLIF(trim(description), ''), content);

ALTER TABLE daily_challenges
  ALTER COLUMN title SET NOT NULL,
  ALTER COLUMN description SET NOT NULL;

CREATE TABLE IF NOT EXISTS daily_physical_challenges (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title text NOT NULL,
  description text NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_daily_physical_challenges_title_description
  ON daily_physical_challenges(title, description);
