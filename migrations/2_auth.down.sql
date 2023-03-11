DROP TABLE IF EXISTS auth;
DROP TRIGGER IF EXISTS tg_auth_set_update_at ON auth RESTRICT;
DROP FUNCTION IF EXISTS tgf_auth_set_updated_at();