-- OJO: usuarios.rol_id tiene ON DELETE CASCADE hacia roles.id. Si para
-- cuando se revierte esta migracion ya existe algun usuario con rol
-- Operario, este DELETE tambien borrara a esos usuarios. Revisar antes
-- de correr esto fuera de un entorno de desarrollo recien creado.
DELETE FROM roles WHERE nombre = 'Operario';

UPDATE roles SET nombre = 'Empleado' WHERE nombre = 'Recepcionista';