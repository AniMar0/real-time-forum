ALTER TABLE messages ADD COLUMN sender_id INTEGER;
ALTER TABLE messages ADD COLUMN receiver_id INTEGER;

UPDATE messages
SET sender_id = (SELECT id FROM users WHERE users.nickname = messages.sender),
    receiver_id = (SELECT id FROM users WHERE users.nickname = messages.receiver);

ALTER TABLE sessions ADD COLUMN user_id INTEGER;

UPDATE sessions
SET user_id = (SELECT id FROM users WHERE users.nickname = sessions.nickname);

ALTER TABLE notifications ADD COLUMN receiver_id INTEGER;
ALTER TABLE notifications ADD COLUMN sender_id INTEGER;

UPDATE notifications
SET receiver_id = (SELECT id FROM users WHERE users.nickname = notifications.receiver_nickname),
    sender_id = (SELECT id FROM users WHERE users.nickname = notifications.sender_nickname);

CREATE INDEX IF NOT EXISTS idx_messages_sender_receiver_id
    ON messages(sender_id, receiver_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_messages_receiver_sender_id
    ON messages(receiver_id, sender_id, id DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id
    ON sessions(user_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_notifications_receiver_sender_id
    ON notifications(receiver_id, sender_id);
