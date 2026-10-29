package main

import (
	"context"
	"log"

	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/config"
	"github.com/ESJ0/selava-backend/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	const email = "admin@selava.local"
	const password = "Admin123!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.NewPostgresPool(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Los roles ya existen para
	// este punto: los crea la migracion 016, igual que estados_pedido se
	// siembra por migracion en vez de por este script. Aqui solo se usa
	// el rol Administrador, que ya deberia estar en la base de datos.
	if _, err = tx.Exec(ctx, `INSERT INTO usuarios(rol_id,nombre,apellido,email,password_hash,telefono) SELECT id,'Admin','SeLava',$1,$2,'00000000' FROM roles WHERE nombre='Administrador' ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash,activo=TRUE,updated_at=NOW()`, email, hash); err != nil {
		log.Fatal(err)
	}
	for _, row := range [][2]string{{"Lavado", "Lavado general"}, {"Planchado", "Planchado de prendas"}, {"Lavado en seco", "Limpieza en seco"}} {
		if _, err = tx.Exec(ctx, `INSERT INTO servicios(nombre,descripcion,tiempo_estimado_horas) VALUES($1,$2,24) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, row[0], row[1]); err != nil {
			log.Fatal(err)
		}
	}
	for _, garment := range []struct {
		name, mode string
	}{{"Camisa", "UNIDAD"}, {"Pantalón", "UNIDAD"}, {"Vestido", "UNIDAD"}, {"Canasta", "LIBRA"}, {"Alfombra", "METRO"}} {
		// Q25 es exclusivamente una tarifa de ejemplo del seed local.
		// Un conflicto conserva la tarifa real que ya haya configurado el administrador.
		if _, err = tx.Exec(ctx, `INSERT INTO tipos_prenda(nombre,precio,modalidad_cobro) VALUES($1,25,$2) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, garment.name, garment.mode); err != nil {
			log.Fatal(err)
		}
	}
	for _, name := range []string{"Efectivo", "Tarjeta", "Transferencia"} {
		if _, err = tx.Exec(ctx, `INSERT INTO metodos_pago(nombre) VALUES($1) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, name); err != nil {
			log.Fatal(err)
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO clientes(nombre,apellido,telefono,email,direccion) VALUES('Cliente','Demo','55550000','cliente.demo@selava.local','Local') ON CONFLICT(email) DO UPDATE SET activo=TRUE`); err != nil {
		log.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	log.Printf("seed de desarrollo aplicado; administrador local: %s", email)
}
