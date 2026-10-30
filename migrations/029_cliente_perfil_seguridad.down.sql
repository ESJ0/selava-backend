ALTER TABLE clientes DROP CONSTRAINT IF EXISTS chk_clientes_foto_perfil_completa;
ALTER TABLE clientes DROP COLUMN IF EXISTS foto_perfil_tipo;
ALTER TABLE clientes DROP COLUMN IF EXISTS foto_perfil;
