DROP TRIGGER IF EXISTS update_post_modtime ON posts;

DROP FUNCTION IF EXISTS update_modified_column();

DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_posts_created_at;

DROP TABLE IF EXISTS posts;