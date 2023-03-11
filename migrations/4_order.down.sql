DROP TABLE IF EXISTS "order";
DROP SEQUENCE IF EXISTS order_base_order_id_seq;
DROP SEQUENCE IF EXISTS order_order_id_seq;
DROP TRIGGER IF EXISTS tg_order_order_id ON "order" RESTRICT;
DROP TRIGGER IF EXISTS tg_order_set_updated_at ON "order" RESTRICT;
DROP FUNCTION IF EXISTS tgf_order_order_id();
DROP FUNCTION IF EXISTS tgf_order_set_updated_at();
