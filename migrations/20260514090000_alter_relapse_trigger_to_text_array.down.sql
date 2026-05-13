ALTER TABLE check_ins
  ALTER COLUMN relapse_trigger TYPE text
  USING CASE
    WHEN relapse_trigger IS NULL OR array_length(relapse_trigger, 1) IS NULL OR array_length(relapse_trigger, 1) = 0 THEN NULL
    ELSE relapse_trigger[1]
  END;
