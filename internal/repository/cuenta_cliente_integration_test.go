package repository

import "testing"

func TestCuentaClienteMantieneRelacionUnoAUno(t *testing.T) {
	f := newSprint3Fixture(t)

	_, err := f.db.Exec(f.ctx, `
		INSERT INTO cuentas_cliente(cliente_id, email, password_hash)
		VALUES($1, 'cliente@selava.test', 'hash-de-prueba')
	`, f.clientID)
	mustSprint3(t, err)

	if _, err = f.db.Exec(f.ctx, `
		INSERT INTO cuentas_cliente(cliente_id, email, password_hash)
		VALUES($1, 'otro@selava.test', 'hash-de-prueba')
	`, f.clientID); err == nil {
		t.Fatal("se permitio crear mas de una cuenta para el mismo cliente")
	}
}

func TestCuentaClienteProtegeCorreoYClienteRelacionado(t *testing.T) {
	f := newSprint3Fixture(t)

	_, err := f.db.Exec(f.ctx, `
		INSERT INTO cuentas_cliente(cliente_id, email, password_hash)
		VALUES($1, 'cliente@selava.test', 'hash-de-prueba')
	`, f.clientID)
	mustSprint3(t, err)

	var otroClienteID int
	err = f.db.QueryRow(f.ctx, `
		INSERT INTO clientes(nombre, apellido, telefono)
		VALUES('Otro', 'Cliente', '11111111')
		RETURNING id
	`).Scan(&otroClienteID)
	mustSprint3(t, err)

	if _, err = f.db.Exec(f.ctx, `
		INSERT INTO cuentas_cliente(cliente_id, email, password_hash)
		VALUES($1, 'cliente@selava.test', 'hash-de-prueba')
	`, otroClienteID); err == nil {
		t.Fatal("se permitio reutilizar el correo de acceso")
	}

	if _, err = f.db.Exec(f.ctx, `DELETE FROM clientes WHERE id=$1`, f.clientID); err == nil {
		t.Fatal("se permitio borrar un cliente que conserva una cuenta")
	}
}
