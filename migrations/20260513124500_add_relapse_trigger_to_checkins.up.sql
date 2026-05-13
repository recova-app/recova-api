ALTER TABLE check_ins
  ADD COLUMN IF NOT EXISTS relapse_trigger text;
