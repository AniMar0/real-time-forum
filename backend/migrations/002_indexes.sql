CREATE INDEX IF NOT EXISTS idx_posts_created_at
    ON posts(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_comments_post_created_at
    ON comments(post_id, created_at ASC);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id
    ON messages(sender, receiver, id DESC);

CREATE INDEX IF NOT EXISTS idx_messages_reverse_conversation_id
    ON messages(receiver, sender, id DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_expiry
    ON sessions(session_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_notifications_pair
    ON notifications(receiver_nickname, sender_nickname);
