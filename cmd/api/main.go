package main

import (
	"log"

	"github.com/ESJ0/selava-backend/internal/config"
	"github.com/ESJ0/selava-backend/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}

	db, err := database.NewPostgresPool(cfg.DSN())
	if err != nil {
		log.Fatalf("error conectando a la base de datos: %v", err)
	}
	defer db.Close()

	log.Println("Conexión a PostgreSQL exitosa")

}
