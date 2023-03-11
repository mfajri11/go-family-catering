CREATE OR REPLACE FUNCTION tgf_auth_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE IF NOT EXISTS "auth"(
    sid TEXT NOT NULL PRIMARY KEY,
	owner_id BIGINT NOT NULL,
	jti TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
	refresh_token TEXT NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER tg_auth_set_update_at
BEFORE UPDATE ON "auth"
FOR EACH ROW
EXECUTE PROCEDURE tgf_auth_set_updated_at();
