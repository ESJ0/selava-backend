ALTER TABLE clientes DROP CONSTRAINT IF EXISTS chk_clientes_nit_formato;
ALTER TABLE clientes DROP COLUMN IF EXISTS nit;
