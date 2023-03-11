CREATE OR REPLACE FUNCTION tgf_owner_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE IF NOT EXISTS "owner"(
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL,
	password VARCHAR(255) NOT NULL,
	phone_number VARCHAR(20) NOT NULL,
	date_of_birth Date NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER tg_owner_set_updated_at
BEFORE UPDATE ON "owner"
FOR EACH ROW
EXECUTE PROCEDURE tgf_owner_set_updated_at();