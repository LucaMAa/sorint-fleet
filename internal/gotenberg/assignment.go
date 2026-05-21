package gotenberg

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"time"
)

//go:embed templates/assignment.html
var assignmentTemplateRaw string
//go:embed templates/logo.png
var logoBytes []byte

var assignmentTmpl = template.Must(
	template.New("assignment").Parse(assignmentTemplateRaw),
)

type AssignmentData struct {
	SignerName   string
	AssigneeName string
	VehicleBrand string
	VehicleModel string
	LicensePlate string
	IsJolly      bool
	Date         time.Time
}

type templateData struct {
	AssignmentData
	DateFormatted   string
	UsagePromiscuo  bool
	UsageLavorativo bool
  LogoBase64      string 
}

type Generator struct {
	client *Client
}

func NewGenerator(client *Client) *Generator {
	return &Generator{client: client}
}

func (g *Generator) GenerateAssignment(data AssignmentData) ([]byte, error) {
	if data.Date.IsZero() {
		data.Date = time.Now()
	}

	htmlBytes, err := renderAssignmentHTML(data)
	if err != nil {
		return nil, fmt.Errorf("pdf: render html: %w", err)
	}

	pdf, err := g.client.HTMLToPDF(htmlBytes, DefaultA4Options)
	if err != nil {
		return nil, fmt.Errorf("pdf: convert to pdf: %w", err)
	}

	return pdf, nil
}

func renderAssignmentHTML(data AssignmentData) ([]byte, error) {
	td := templateData{
		AssignmentData:  data,
		DateFormatted:   data.Date.Format("02/01/2006"),
		UsagePromiscuo:  !data.IsJolly,
		UsageLavorativo: data.IsJolly,
    LogoBase64: base64.StdEncoding.EncodeToString(logoBytes),
	}
	var buf bytes.Buffer
	if err := assignmentTmpl.Execute(&buf, td); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
