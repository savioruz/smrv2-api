BEGIN;

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nim VARCHAR(20) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) DEFAULT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    major VARCHAR(255) DEFAULT NULL,
    level CHAR(1) NOT NULL DEFAULT '1',
    last_login TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    reset_password_token VARCHAR(255) DEFAULT NULL,
    verification_token VARCHAR(255) DEFAULT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_portal_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    plan_type VARCHAR(20) NOT NULL, -- 'free', 'basic', 'premium'
    status VARCHAR(20) NOT NULL, -- 'active', 'expired', 'cancelled'
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_users_nim ON users(nim);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users USING btree (deleted_at ASC NULLS LAST);
CREATE INDEX idx_subscriptions_deleted_at ON subscriptions USING btree (deleted_at ASC NULLS LAST);

COMMIT;
