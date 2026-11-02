-- Constancia de aceptación de los Términos y condiciones al crear la cuenta de cliente.
ALTER TABLE usuarios
    ADD COLUMN terminos_aceptados_at TIMESTAMP,
    ADD COLUMN terminos_version VARCHAR(20);
