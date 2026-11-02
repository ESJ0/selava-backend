-- Email y dirección vacíos pasan a NULL. clientes.email es UNIQUE: con '' solo
-- podía existir un cliente sin correo y el segundo fallaba al guardarse.
UPDATE clientes SET email = NULL WHERE email IS NOT NULL AND BTRIM(email) = '';
UPDATE clientes SET direccion = NULL WHERE direccion IS NOT NULL AND BTRIM(direccion) = '';
