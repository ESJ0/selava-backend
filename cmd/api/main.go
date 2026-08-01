package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ESJ0/selava-backend/internal/config"
	"github.com/ESJ0/selava-backend/internal/controller"
	"github.com/ESJ0/selava-backend/internal/database"
	"github.com/ESJ0/selava-backend/internal/repository"
	"github.com/ESJ0/selava-backend/internal/routes"
	"github.com/ESJ0/selava-backend/internal/service"
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

	clienteRepo := repository.NewClienteRepository(db)
	clienteService := service.NewClienteService(clienteRepo)
	clienteController := controller.NewClienteController(clienteService)

	router := routes.NewRouter(clienteController)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Servidor escuchando en el puerto %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error iniciando el servidor: %v", err)
		}
	}()

	// Apagado controlado: si se reinicia el servicio (deploy, restart de
	// contenedor), termina las peticiones en curso en vez de cortarlas
	// a medio ingresar una venta.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Apagando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("error en apagado controlado: %v", err)
	}
	log.Println("Servidor apagado correctamente")
}
