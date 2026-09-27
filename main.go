package main

import (
	"log"
	"net/http"

	"github.com/gweppi/pdf-gen/config"
	"github.com/gweppi/pdf-gen/server"
)

func main() {
	cfg := config.Load()

	srv := server.NewServer(&cfg)

	// Add API Key auth if env var set
	renderHandler := srv.RenderPDFHandler
	if cfg.ApiKey != "" {
		renderHandler = server.APIKeyMiddleware(cfg.ApiKey, srv.RenderPDFHandler)
	}

	http.HandleFunc("/render/", renderHandler)

	// Set path to access generated documents
	generatedHandler := srv.ServeGeneratedPDFHandler
	// if cfg.ApiKey != "" {
	// 	generatedHandler = server.APIKeyMiddleware(cfg.ApiKey, srv.ServeGeneratedPDFHandler)
	// }

	http.HandleFunc("/generated/", generatedHandler)

	log.Printf("PDF Rendering Server running on %s", cfg.Hostname)
	log.Printf("Send POST requests to %srender/<template_name>", cfg.Hostname)
	log.Printf("Send GET requests to %sgenerate/<file_name>.pdf", cfg.Hostname)

	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
