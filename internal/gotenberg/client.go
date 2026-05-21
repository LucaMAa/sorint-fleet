package gotenberg

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// Client handles communication with a Gotenberg instance.
type Client struct {
	BaseURL string
}

// HTMLToPDFOptions controls Chromium rendering parameters.
type HTMLToPDFOptions struct {
	PaperWidth      string
	PaperHeight     string
	MarginTop       string
	MarginBottom    string
	MarginLeft      string
	MarginRight     string
	PrintBackground bool
	Scale           string
}

// DefaultA4Options returns standard A4 options with no margins.
var DefaultA4Options = HTMLToPDFOptions{
	PaperWidth:      "8.27",
	PaperHeight:     "11.69",
	MarginTop:       "0",
	MarginBottom:    "0",
	MarginLeft:      "0",
	MarginRight:     "0",
	PrintBackground: true,
	Scale:           "1.0",
}

// NewClient creates a Client, falling back to $GOTENBERG_URL or localhost:3000.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = os.Getenv("GOTENBERG_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	return &Client{BaseURL: baseURL}
}

// HTMLToPDF converts raw HTML bytes to a PDF using Gotenberg's Chromium endpoint.
func (c *Client) HTMLToPDF(htmlContent []byte, opts HTMLToPDFOptions) ([]byte, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	fw, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("gotenberg: create form file: %w", err)
	}
	if _, err = fw.Write(htmlContent); err != nil {
		return nil, fmt.Errorf("gotenberg: write html: %w", err)
	}

	fields := map[string]string{
		"paperWidth":      opts.PaperWidth,
		"paperHeight":     opts.PaperHeight,
		"marginTop":       opts.MarginTop,
		"marginBottom":    opts.MarginBottom,
		"marginLeft":      opts.MarginLeft,
		"marginRight":     opts.MarginRight,
		"scale":           opts.Scale,
	}
	if opts.PrintBackground {
		fields["printBackground"] = "true"
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("gotenberg: write field %q: %w", k, err)
		}
	}
	w.Close()

	resp, err := http.Post(
		c.BaseURL+"/forms/chromium/convert/html",
		w.FormDataContentType(),
		&body,
	)
	if err != nil {
		return nil, fmt.Errorf("gotenberg: unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gotenberg: status %d: %s", resp.StatusCode, string(b))
	}

	return io.ReadAll(resp.Body)
}
