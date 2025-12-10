// Package templates provides email template rendering functionality.
package templates

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

//go:embed *.html *.txt
var templateFS embed.FS

// Renderer handles email template rendering.
type Renderer struct {
	htmlTemplates *htmltemplate.Template
	textTemplates *texttemplate.Template
}

// NewRenderer creates a new template renderer.
func NewRenderer() (*Renderer, error) {
	htmlTmpl, err := htmltemplate.ParseFS(templateFS, "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML templates: %w", err)
	}

	textTmpl, err := texttemplate.ParseFS(templateFS, "*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to parse text templates: %w", err)
	}

	return &Renderer{
		htmlTemplates: htmlTmpl,
		textTemplates: textTmpl,
	}, nil
}

// Render renders both HTML and text versions of a template.
func (r *Renderer) Render(templateName string, data interface{}) (html string, text string, err error) {
	// Render HTML
	var htmlBuf bytes.Buffer
	if err := r.htmlTemplates.ExecuteTemplate(&htmlBuf, templateName+".html", data); err != nil {
		return "", "", fmt.Errorf("failed to render HTML template %s: %w", templateName, err)
	}

	// Render text
	var textBuf bytes.Buffer
	if err := r.textTemplates.ExecuteTemplate(&textBuf, templateName+".txt", data); err != nil {
		// Fall back to empty text if no text template exists
		return htmlBuf.String(), "", nil
	}

	return htmlBuf.String(), textBuf.String(), nil
}

// RenderHTML renders only the HTML version of a template.
func (r *Renderer) RenderHTML(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := r.htmlTemplates.ExecuteTemplate(&buf, templateName+".html", data); err != nil {
		return "", fmt.Errorf("failed to render HTML template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// RenderText renders only the text version of a template.
func (r *Renderer) RenderText(templateName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := r.textTemplates.ExecuteTemplate(&buf, templateName+".txt", data); err != nil {
		return "", fmt.Errorf("failed to render text template %s: %w", templateName, err)
	}
	return buf.String(), nil
}

// PasswordResetData contains data for password reset email template.
type PasswordResetData struct {
	UserName  string
	ResetURL  string
	ExpiresIn string
}

// GroupInvitationData contains data for group invitation email template.
type GroupInvitationData struct {
	InviterName  string
	InviterEmail string
	GroupName    string
	InviteURL    string
	ExpiresIn    string
}
