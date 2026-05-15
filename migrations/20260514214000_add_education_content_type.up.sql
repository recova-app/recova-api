ALTER TABLE education_contents
  ADD COLUMN IF NOT EXISTS type text;

UPDATE education_contents
SET type = 'artikel'
WHERE type IS NULL
  OR btrim(type) = ''
  OR type NOT IN ('artikel', 'video');

ALTER TABLE education_contents
  ALTER COLUMN type SET DEFAULT 'artikel',
  ALTER COLUMN type SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'education_contents_type_check'
  ) THEN
    ALTER TABLE education_contents
      ADD CONSTRAINT education_contents_type_check
      CHECK (type IN ('artikel', 'video'));
  END IF;
END $$;
