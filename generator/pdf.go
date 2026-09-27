package generator

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

type PDFGenerator struct {
	templatesDir string
	gotenbergUrl string
}

func NewPDFGenerator(templatesDir, gotenbergUrl string) *PDFGenerator {
	return &PDFGenerator{
		templatesDir: templatesDir,
		gotenbergUrl: gotenbergUrl,
	}
}

// Compile the regex once at startup for optimal performance
var validTemplateRegex = regexp.MustCompile(`^[a-z0-9_]+$`)

func (g *PDFGenerator) Generate(ctx context.Context, templateName string, data map[string]any) (io.ReadCloser, error) {

	if !validTemplateRegex.MatchString(templateName) {
		return nil, fmt.Errorf("Invalid template name. Only lowercase letters (a-z), numbers (0-9), and underscores (_)  are allowed (e.g., test_1).")
	}

	templatePath := filepath.Join(g.templatesDir, templateName+".html")

	tmpl, err := template.New(filepath.Base(templatePath)).
		ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("Template not found or invalid: %s", templateName)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := tmpl.Execute(pw, data); err != nil {
			pw.CloseWithError(err)
		}
	}()

	doc, err := document.FromReader("index.html", pr)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize document stream")
	}

	req := gotenberg.NewHTMLRequest(doc)
	req.PaperSize(gotenberg.A4)
	req.PrintBackground()

	gotenberg_url := g.gotenbergUrl
	client, err := gotenberg.NewClient(gotenberg_url, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("PDF Engine unavailable")
	}

	res, err := client.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate PDF")
	}

	return res.Body, nil
}
