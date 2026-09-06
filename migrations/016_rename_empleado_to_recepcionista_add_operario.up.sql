-- Deja los 3 roles finales del sistema en su lugar, sin importar si esta
-- migracion corre sobre una base de datos nueva (con la tabla roles vacia)
-- o sobre una base de datos existente que ya tenia los roles viejos
-- "Administrador" y "Empleado" sembrados por cmd/seed.
--
-- Si existe "Empleado", se renombra a "Recepcionista" conservando su id
-- para no romper la referencia de los usuarios que ya lo tenian asignado.
UPDATE roles SET nombre = 'Recepcionista' WHERE nombre = 'Empleado';

-- Estos INSERT con ON CONFLICT DO NOTHING no duplican roles existentes.
-- En una base de datos nueva garantizan el orden esperado de ids:
-- Administrador=1, Recepcionista=2, Operario=3.
INSERT INTO roles (nombre) VALUES ('Administrador') ON CONFLICT (nombre) DO NOTHING;
INSERT INTO roles (nombre) VALUES ('Recepcionista') ON CONFLICT (nombre) DO NOTHING;
INSERT INTO roles (nombre) VALUES ('Operario')      ON CONFLICT (nombre) DO NOTHING;
