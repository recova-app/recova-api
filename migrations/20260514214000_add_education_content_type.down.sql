ALTER TABLE education_contents
  DROP CONSTRAINT IF EXISTS education_contents_type_check;

ALTER TABLE education_contents
  ALTER COLUMN type DROP NOT NULL,
  ALTER COLUMN type DROP DEFAULT;

ALTER TABLE education_contents
  DROP COLUMN IF EXISTS type;
