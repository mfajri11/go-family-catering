DROP TABLE IF EXISTS menu;
DROP SEQUENCE IF EXISTS menu_id_seq;
DROP TRIGGER IF EXISTS tg_menu_set_updated_at ON menu RESTRICT;
DROP FUNCTION IF EXISTS tgf_menu_set_updated_at();