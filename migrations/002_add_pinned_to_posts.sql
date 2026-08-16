-- Add pinned column to posts table
ALTER TABLE posts ADD COLUMN pinned BOOLEAN DEFAULT 0;

-- Create index for pinned posts
CREATE INDEX idx_posts_pinned ON posts(pinned);
