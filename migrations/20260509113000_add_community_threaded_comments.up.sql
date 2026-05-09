ALTER TABLE community_comments
  ADD COLUMN IF NOT EXISTS parent_comment_id uuid,
  ADD COLUMN IF NOT EXISTS depth smallint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS reply_count integer NOT NULL DEFAULT 0;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_community_comments_parent_comment'
  ) THEN
    ALTER TABLE community_comments
      ADD CONSTRAINT fk_community_comments_parent_comment
      FOREIGN KEY (parent_comment_id)
      REFERENCES community_comments(id)
      ON UPDATE CASCADE
      ON DELETE CASCADE;
  END IF;
END;
$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_community_comments_depth_non_negative'
  ) THEN
    ALTER TABLE community_comments
      ADD CONSTRAINT chk_community_comments_depth_non_negative
      CHECK (depth >= 0);
  END IF;
END;
$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_community_comments_parent_depth_consistency'
  ) THEN
    ALTER TABLE community_comments
      ADD CONSTRAINT chk_community_comments_parent_depth_consistency
      CHECK (
        (parent_comment_id IS NULL AND depth = 0)
        OR (parent_comment_id IS NOT NULL AND depth > 0)
      );
  END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_community_comments_parent
  ON community_comments(parent_comment_id);

CREATE INDEX IF NOT EXISTS idx_community_comments_post_parent_created
  ON community_comments(post_id, parent_comment_id, created_at, id);
