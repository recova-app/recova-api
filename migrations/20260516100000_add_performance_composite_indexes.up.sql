CREATE INDEX IF NOT EXISTS idx_journals_user_created
  ON journals(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_community_posts_created_at_desc
  ON community_posts(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_community_posts_category_created_at_desc
  ON community_posts(category, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_relapses_user_created
  ON relapses(user_id, created_at DESC);
