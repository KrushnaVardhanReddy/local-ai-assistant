package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// DocumentParser defines the interface for parsing text from a document.
type DocumentParser interface {
	Parse(filepath string) (string, error)
}

// TextParser parses .txt and .md files.
type TextParser struct{}

func (p *TextParser) Parse(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PDFParser parses .pdf files.
type PDFParser struct{}

func (p *PDFParser) Parse(filePath string) (string, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	b, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)

	return buf.String(), nil
}

// PPTXParser parses .pptx files by reading the zipped slide XMLs.
type PPTXParser struct{}

func (p *PPTXParser) Parse(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var textBuilder strings.Builder

	for _, f := range r.File {
		// Only look for slide XMLs in ppt/slides/
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			// ZipSlip prevention check
			if !strings.HasPrefix(filepath.Clean(f.Name), "ppt/slides/slide") {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			// Use XML decoding to extract all text nodes
			decoder := xml.NewDecoder(rc)
			for {
				t, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					rc.Close()
					return "", err
				}

				switch se := t.(type) {
				case xml.CharData:
					str := strings.TrimSpace(string(se))
					if str != "" {
						textBuilder.WriteString(str + " ")
					}
				}
			}
			textBuilder.WriteString("\n")
			rc.Close()
		}
	}

	return strings.TrimSpace(textBuilder.String()), nil
}

// Registry manages available parsers by file extension.
type Registry struct {
	parsers map[string]DocumentParser
}

// NewRegistry creates a new parser registry with default parsers.
func NewRegistry() *Registry {
	return &Registry{
		parsers: map[string]DocumentParser{
			".txt":  &TextParser{},
			".md":   &TextParser{},
			".pdf":  &PDFParser{},
			".pptx": &PPTXParser{},
		},
	}
}

// GetParser returns the appropriate parser for the given filepath extension.
func (r *Registry) GetParser(filepath string) (DocumentParser, error) {
	ext := strings.ToLower(filepath[strings.LastIndex(filepath, "."):])
	if parser, exists := r.parsers[ext]; exists {
		return parser, nil
	}
	return nil, fmt.Errorf("no parser found for extension: %s", ext)
}

// ParseDocument is a convenience function to parse a file using the registry.
func ParseDocument(filePath string) (string, error) {
	registry := NewRegistry()
	parser, err := registry.GetParser(filePath)
	if err != nil {
		return "", err
	}
	return parser.Parse(filePath)
}
