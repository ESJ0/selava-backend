package main

import (
	"context"
	"github.com/ESJ0/selava-backend/internal/auth"
	"github.com/ESJ0/selava-backend/internal/config"
	"github.com/ESJ0/selava-backend/internal/database"
	"log"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	email, password := os.Getenv("SEED_ADMIN_EMAIL"), os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("SEED_ADMIN_EMAIL y SEED_ADMIN_PASSWORD son requeridos")
	}
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
	for _, role := range []string{"Administrador", "Empleado"} {
		if _, err = tx.Exec(ctx, `INSERT INTO roles(nombre) VALUES($1) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, role); err != nil {
			log.Fatal(err)
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO usuarios(rol_id,nombre,apellido,email,password_hash,telefono) SELECT id,'Admin','SeLava',$1,$2,'00000000' FROM roles WHERE nombre='Administrador' ON CONFLICT(email) DO UPDATE SET password_hash=EXCLUDED.password_hash,activo=TRUE`, email, hash); err != nil {
		log.Fatal(err)
	}
	for _, row := range [][2]string{{"Lavado", "Lavado general"}, {"Planchado", "Planchado de prendas"}, {"Lavado en seco", "Limpieza en seco"}} {
		if _, err = tx.Exec(ctx, `INSERT INTO servicios(nombre,descripcion,precio_base,tiempo_estimado_horas) VALUES($1,$2,25,24) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, row[0], row[1]); err != nil {
			log.Fatal(err)
		}
	}
	for _, name := range []string{"Camisa", "Pantalón", "Vestido"} {
		if _, err = tx.Exec(ctx, `INSERT INTO tipos_prenda(nombre) VALUES($1) ON CONFLICT(nombre) DO UPDATE SET activo=TRUE`, name); err != nil {
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
	log.Println("seed de desarrollo aplicado")
}
