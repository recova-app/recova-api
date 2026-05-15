ALTER TABLE check_ins
  ALTER COLUMN relapse_trigger TYPE text[]
  USING CASE
    WHEN relapse_trigger IS NULL OR btrim(relapse_trigger) = '' THEN NULL
    ELSE ARRAY[relapse_trigger]
  END;
