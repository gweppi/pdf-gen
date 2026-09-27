package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"uuid"

	"github.com/gweppi/pdf-gen/config"
	"github.com/gweppi/pdf-gen/generator"
)

type Server struct {
	Config       config.Config
	pdfGenerator *generator.PDFGenerator
}

type RenderPDFResponse struct {
	FileName string `json:"filename"`
	Url      string `json:"url"`
}

func NewServer(config *config.Config) *Server {
	return &Server{
		Config:       *config,
		pdfGenerator: generator.NewPDFGenerator(config.TemplatesDir, config.GotenbergUrl),
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

	// Write data to disk
	id := uuid.New().String()
	fileName := id + ".pdf"

	fullPath := filepath.Join(s.Config.OutputDir, fileName)

	dst, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		http.Error(w, "Failed to store file on server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, pdfStream); err != nil {
		// If streaming fails, clean up the partial file
		os.Remove(fullPath)
		http.Error(w, "Failed to write file to disk", http.StatusInternalServerError)
		return
	}

	hostPath := s.Config.Hostname + "generated/" + fileName

	responseType := r.URL.Query().Get("response")
	switch responseType {
	case "link":
		resp := RenderPDFResponse{
			FileName: fileName,
			Url:      hostPath,
		}
		fmt.Println(resp)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "redirect":
		http.Redirect(w, r, hostPath, http.StatusFound)
	case "binary":
	default:
		// 2. Open the newly saved file from disk to stream to the client
		savedFile, err := os.Open(filepath.Join(s.Config.OutputDir, fileName))
		if err != nil {
			http.Error(w, "Failed to open saved PDF", http.StatusInternalServerError)
			return
		}
		defer savedFile.Close() // Clean up file descriptor when finished

		// 3. Stream the file directly to the HTTP response
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=\""+fileName+"\"")
		io.Copy(w, savedFile)
	}

	log.Printf("Successfully generated pdf from template named: %s", templateName)
}

func (s *Server) ServeGeneratedPDFHandler(w http.ResponseWriter, r *http.Request) {
	documentName := strings.TrimPrefix(r.URL.Path, "/generated/")
	// Take out things like path traversals etc.
	cleanDocumentName := filepath.Base(documentName)

	fullPath := filepath.Join(s.Config.OutputDir, cleanDocumentName)

	_, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\""+cleanDocumentName+"\"")
	http.ServeFile(w, r, fullPath)
}
