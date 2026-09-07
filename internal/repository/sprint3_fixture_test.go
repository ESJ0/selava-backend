package repository

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ESJ0/selava-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Cada test usa un esquema nuevo y las migraciones reales. Nunca usa .env
// ni la base de desarrollo. El nombre de la DB debe terminar en _test.
type sprint3Fixture struct {
	db *pgxpool.Pool
	ctx context.Context
	userID, clientID, typeID int
	services []int
}

func newSprint3Fixture(t *testing.T) *sprint3Fixture {
	t.Helper()
	dsn := os.Getenv("SELAVA_TEST_DATABASE_URL")
	if dsn == "" { t.Skip("requiere SELAVA_TEST_DATABASE_URL para pruebas de PostgreSQL") }
	config, err := pgxpool.ParseConfig(dsn)
	mustSprint3(t, err)
	if !strings.HasSuffix(config.ConnConfig.Database, "_test") { t.Fatal("la base de pruebas debe terminar en _test") }
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, dsn)
	mustSprint3(t, err)
	schema := pgx.Identifier{fmt.Sprintf("sprint3_%d", time.Now().UnixNano())}.Sanitize()
	_, err = admin.Exec(ctx, "CREATE SCHEMA " + schema)
	mustSprint3(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema
	config.MaxConns = 4
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil { _, _ = admin.Exec(ctx, "DROP SCHEMA " + schema + " CASCADE"); _ = admin.Close(ctx); t.Fatal(err) }
	t.Cleanup(func() {
		db.Close()
		_, cleanupErr := admin.Exec(ctx, "DROP SCHEMA " + schema + " CASCADE")
		if cleanupErr != nil { t.Errorf("limpiando esquema de test: %v", cleanupErr) }
		_ = admin.Close(ctx)
	})
	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	mustSprint3(t, err)
	if len(files) == 0 { t.Fatal("no se encontraron migraciones") }
	for _, file := range files {
		sql, readErr := os.ReadFile(file); mustSprint3(t, readErr)
		if _, err = db.Exec(ctx, string(sql)); err != nil { t.Fatalf("migracion %s: %v", file, err) }
	}
	f := &sprint3Fixture{db: db, ctx: ctx}
	mustSprint3(t, db.QueryRow(ctx, `INSERT INTO usuarios(rol_id,nombre,apellido,email,password_hash,telefono) VALUES(1,'Usuario','Prueba','sprint3@test.local','test-only','00000000') RETURNING id`).Scan(&f.userID))
	mustSprint3(t, db.QueryRow(ctx, `INSERT INTO clientes(nombre,apellido,telefono) VALUES('Cliente','Prueba','00000000') RETURNING id`).Scan(&f.clientID))
	mustSprint3(t, db.QueryRow(ctx, `INSERT INTO tipos_prenda(nombre) VALUES('Camisa') RETURNING id`).Scan(&f.typeID))
	for i, price := range []float64{10.25, 5.50, 2} {
		var id int
		mustSprint3(t, db.QueryRow(ctx, `INSERT INTO servicios(nombre,precio_base) VALUES($1,$2) RETURNING id`, fmt.Sprintf("Servicio %d", i), price).Scan(&id))
		f.services = append(f.services, id)
	}
	return f
}

func mustSprint3(t *testing.T, err error) { t.Helper(); if err != nil { t.Fatal(err) } }

func (f *sprint3Fixture) order(t *testing.T, quantities ...int) *models.PedidoConPrendas {
	t.Helper()
	if len(quantities) == 0 { quantities = []int{2} }
	req := &models.PedidoCreateRequest{ClienteID: f.clientID}
	for _, quantity := range quantities {
		req.Prendas = append(req.Prendas, models.PrendaCreateRequest{TipoPrendaID: f.typeID, Cantidad: quantity, Servicios: []models.PrendaServicioCreateRequest{{ServicioID: f.services[0]}}})
	}
	order, err := NewPedidoRepository(f.db).Create(f.ctx, req, f.userID)
	mustSprint3(t, err)
	return order
}

func (f *sprint3Fixture) total(t *testing.T, orderID int, want float64) {
	t.Helper()
	var total float64
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT total FROM pedidos WHERE id=$1`, orderID).Scan(&total))
	if math.Abs(total-want) > 0.00001 { t.Fatalf("total=%v, esperado=%v", total, want) }
}

func (f *sprint3Fixture) state(t *testing.T, name string) int {
	t.Helper(); var id int
	mustSprint3(t, f.db.QueryRow(f.ctx, `SELECT id FROM estados_pedido WHERE nombre=$1`, name).Scan(&id))
	return id
}
