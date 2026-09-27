package main

import (
	"log"
	"net/http"

	"github.com/gweppi/pdf-gen/generator"
	"github.com/gweppi/pdf-gen/server"
)

func main() {
	config := LoadConfig()

	pdfGenerator := generator.NewPDFGenerator(config.templatesDir, config.gotenbergUrl)
	srv := server.NewServer(pdfGenerator)

	// Add API Key auth if env var set
	handler := srv.RenderPDFHandler
	if config.apiKey != "" {
		handler = server.APIKeyMiddleware(config.apiKey, srv.RenderPDFHandler)
	}

	http.HandleFunc("/render/", handler)

	log.Printf("PDF Rendering Server running on http://localhost:%s", config.port)
	log.Printf("Send POST requests to http://localhost:%s/render/<template_name>", config.port)

	if err := http.ListenAndServe(":"+config.port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
