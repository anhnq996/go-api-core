CREATE TABLE IF NOT EXISTS friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    friend_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT check_different_users CHECK (user_id != friend_id)
);

CREATE INDEX idx_friendships_user1_id ON friendships(user_id);
CREATE INDEX idx_friendships_user2_id ON friendships(friend_id);
CREATE INDEX idx_friendships_deleted_at ON friendships(deleted_at);
CREATE UNIQUE INDEX idx_friendships_unique ON friendships(user_id, friend_id)
WHERE deleted_at IS NULL;

