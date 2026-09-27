package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gweppi/pdf-gen/generator"
)

type Server struct {
	pdfGenerator *generator.PDFGenerator
}

func NewServer(pdfGenerator *generator.PDFGenerator) *Server {
	return &Server{
		pdfGenerator: pdfGenerator,
	}
}

func (s *Server) RenderPDFHandler(w http.ResponseWriter, r *http.Request) {
	templateName := strings.TrimPrefix(r.URL.Path, "/render/")

	var data map[string]any
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber() // Preserves numeric precision without forcing float64
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON body: %v", err), http.StatusBadRequest)
		return
	}

	pdfStream, err := s.pdfGenerator.Generate(r.Context(), templateName, data)
	if err != nil {
		log.Printf("PDF generation error: %v", err)
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
	defer pdfStream.Close()

	responseType := r.URL.Query().Get("response")
	switch responseType {
	case "link":
	case "redirect":
	case "binary":
	default:
		// Stream file as response
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.pdf\"", templateName))

		// Stream binary bytes directly from Gotenberg to the client HTTP response
		if _, err := io.Copy(w, pdfStream); err != nil {
			log.Printf("Error streaming PDF to client: %v", err)
		}
	}

	log.Printf("Successfully generated pdf from template named: %s", templateName)
}
