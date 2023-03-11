DROP TABLE IF EXISTS owner;
DROP SEQUENCE IF EXISTS owner_id_seq;
DROP TRIGGER IF EXISTS tg_owner_set_updated_at ON owner RESTRICT;
DROP FUNCTION IF EXISTS tgf_owner_set_updated_at();