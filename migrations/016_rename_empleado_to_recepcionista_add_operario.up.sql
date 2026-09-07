BEGIN;
-- IDs requeridos por middleware y frontend. ON CONFLICT con SERIAL
-- consumiria IDs y asignaria Operario=5 al actualizar Sprint 2.
UPDATE roles SET nombre = 'Recepcionista' WHERE nombre = 'Empleado';
INSERT INTO roles (id, nombre) VALUES
    (1, 'Administrador'), (2, 'Recepcionista'), (3, 'Operario')
ON CONFLICT (nombre) DO NOTHING;
DO $$
BEGIN
    IF (SELECT id FROM roles WHERE nombre = 'Administrador') <> 1
       OR (SELECT id FROM roles WHERE nombre = 'Recepcionista') <> 2
       OR (SELECT id FROM roles WHERE nombre = 'Operario') <> 3 THEN
        RAISE EXCEPTION 'IDs de roles incompatibles con autorizacion; revisar antes de migrar';
    END IF;
END $$;
SELECT setval(pg_get_serial_sequence('roles', 'id'), (SELECT MAX(id) FROM roles));
COMMIT;
