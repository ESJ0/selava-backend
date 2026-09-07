BEGIN;
-- usuarios.rol_id tiene ON DELETE CASCADE. No borrar usuarios al revertir.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM usuarios u JOIN roles r ON r.id = u.rol_id
               WHERE r.nombre = 'Operario') THEN
        RAISE EXCEPTION 'No se puede revertir: existen usuarios con rol Operario';
    END IF;
END $$;
DELETE FROM roles WHERE nombre = 'Operario';
UPDATE roles SET nombre = 'Empleado' WHERE nombre = 'Recepcionista';
COMMIT;
