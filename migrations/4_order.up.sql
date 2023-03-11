CREATE SEQUENCE IF NOT EXISTS order_order_id_seq;

CREATE OR REPLACE FUNCTION tgf_order_order_id() 
RETURNS TRIGGER AS $$
BEGIN
    PERFORM nextval('order_order_id_seq');
    RETURN NULL;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE OR REPLACE FUNCTION tgf_order_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql VOLATILE;

CREATE TABLE IF NOT EXISTS "order"(
    base_order_id BIGSERIAL NOT NULL PRIMARY KEY,
	customer_email VARCHAR(255) NOT NULL,
	menu_id BIGINT NOT NULL,
	menu_name VARCHAR(255) NOT NULL,
    order_id BIGINT NOT NULL DEFAULT CURRVAL('order_order_id_seq'),
	price FLOAT4 NOT NULL CHECK (price > 0.05),
    qty FLOAT4 NOT NULL DEFAULT 1 CHECK (qty > 0),
	status INT4 NOT NULL DEFAULT 1 CHECK (status > 0 AND status < 4),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER tg_order_order_id
BEFORE INSERT ON "order"
FOR EACH STATEMENT EXECUTE PROCEDURE tgf_order_order_id();

CREATE TRIGGER tg_order_set_updated_at
BEFORE UPDATE ON "order"
FOR EACH ROW
EXECUTE PROCEDURE tgf_order_set_updated_at();
