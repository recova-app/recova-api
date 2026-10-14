CREATE INDEX IF NOT EXISTS idx_streaks_user_active_start_date_desc
  ON streaks(user_id, start_date DESC)
  WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_community_posts_created_at_id_desc
  ON community_posts(created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_community_posts_category_created_at_id_desc
  ON community_posts(category, created_at DESC, id DESC);
