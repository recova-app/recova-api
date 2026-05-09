DROP INDEX IF EXISTS idx_community_comments_post_parent_created;
DROP INDEX IF EXISTS idx_community_comments_parent;

ALTER TABLE community_comments
  DROP CONSTRAINT IF EXISTS chk_community_comments_parent_depth_consistency,
  DROP CONSTRAINT IF EXISTS chk_community_comments_depth_non_negative,
  DROP CONSTRAINT IF EXISTS fk_community_comments_parent_comment;

ALTER TABLE community_comments
  DROP COLUMN IF EXISTS reply_count,
  DROP COLUMN IF EXISTS depth,
  DROP COLUMN IF EXISTS parent_comment_id;
