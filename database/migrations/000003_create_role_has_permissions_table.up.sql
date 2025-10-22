CREATE TABLE IF NOT EXISTS role_has_permissions (
    role_id UUID NOT NULL,
    permission_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_role_has_permissions_role_id ON role_has_permissions(role_id);
CREATE INDEX idx_role_has_permissions_permission_id ON role_has_permissions(permission_id);

