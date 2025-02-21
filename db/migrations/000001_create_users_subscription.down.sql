BEGIN;

DROP TABLE users;
DROP TABLE subscriptions;

DROP INDEX idx_subscriptions_user_id;
DROP INDEX idx_users_nim;

COMMIT;
