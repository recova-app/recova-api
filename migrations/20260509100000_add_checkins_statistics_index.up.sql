CREATE INDEX IF NOT EXISTS idx_check_ins_user_date_success
  ON check_ins(user_id, check_in_date, is_successful);
