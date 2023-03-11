CREATE OR REPLACE FUNCTION tgf_menu_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE IF NOT EXISTS menu(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    categories VARCHAR(255),
    price FLOAT4 NOT NULL CHECK (price > 0.05),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER tg_menu_set_updated_at
BEFORE UPDATE ON "menu"
FOR EACH ROW
EXECUTE PROCEDURE tgf_menu_set_updated_at();
