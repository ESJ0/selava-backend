-- NIT opcional del cliente para facturación. Se guarda normalizado ("1234567-8" o "CF").
ALTER TABLE clientes ADD COLUMN nit VARCHAR(20);
ALTER TABLE clientes ADD CONSTRAINT chk_clientes_nit_formato CHECK (nit IS NULL OR nit ~ '^(CF|[0-9]{1,12}-[0-9K])$');
