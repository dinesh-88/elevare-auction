package emailtpl

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	"strings"
	texttemplate "text/template"

	"github.com/dinesh-88/elevare-auction/internal/domain"
)

//go:embed templates/**
var templateFS embed.FS

type RenderedEmail struct {
	Subject string
	HTML    string
	Text    string
}

type BrandContext struct {
	TenantName    string
	TenantLogoURL string
	PrimaryURL    string
	SupportEmail  string
}

type TemplateRenderer struct {
	fs              fs.FS
	subjectRegistry map[string]map[domain.Locale]string
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		fs: templateFS,
		subjectRegistry: map[string]map[domain.Locale]string{
			"verify_email": {
				domain.LocaleEnglish: "Verify your email for {{.tenant_name}}",
				domain.LocaleFrench:  "Verifiez votre e-mail pour {{.tenant_name}}",
				domain.LocaleGerman:  "Bestaetigen Sie Ihre E-Mail fuer {{.tenant_name}}",
			},
			"password_reset": {
				domain.LocaleEnglish: "Reset your password for {{.tenant_name}}",
				domain.LocaleFrench:  "Reinitialisez votre mot de passe pour {{.tenant_name}}",
				domain.LocaleGerman:  "Setzen Sie Ihr Passwort fuer {{.tenant_name}} zurueck",
			},
			"outbid": {
				domain.LocaleEnglish: "You have been outbid",
				domain.LocaleFrench:  "Vous avez ete surenchéri",
				domain.LocaleGerman:  "Sie wurden ueberboten",
			},
			"lot_won": {
				domain.LocaleEnglish: "You won the lot",
				domain.LocaleFrench:  "Vous avez remporte le lot",
				domain.LocaleGerman:  "Sie haben das Los gewonnen",
			},
			"invoice_ready": {
				domain.LocaleEnglish: "Invoice ready",
				domain.LocaleFrench:  "Facture prete",
				domain.LocaleGerman:  "Rechnung bereit",
			},
		},
	}
}

func (r *TemplateRenderer) Render(templateKey string, locale domain.Locale, brand BrandContext, data map[string]any) (RenderedEmail, error) {
	templateKey = strings.TrimSpace(templateKey)
	if templateKey == "" {
		return RenderedEmail{}, fmt.Errorf("template key is required")
	}
	loc := domain.NormalizeLocale(string(locale))
	ctx := buildTemplateContext(brand, data)

	subjectTemplate, err := r.subjectTemplate(templateKey, loc)
	if err != nil {
		return RenderedEmail{}, err
	}
	subject, err := renderTextString(subjectTemplate, ctx)
	if err != nil {
		return RenderedEmail{}, err
	}

	htmlBody, err := r.renderHTML(templateKey, loc, ctx)
	if err != nil {
		return RenderedEmail{}, err
	}
	textBody, err := r.renderText(templateKey, loc, ctx)
	if err != nil {
		return RenderedEmail{}, err
	}
	return RenderedEmail{
		Subject: strings.TrimSpace(subject),
		HTML:    htmlBody,
		Text:    textBody,
	}, nil
}

func (r *TemplateRenderer) subjectTemplate(templateKey string, locale domain.Locale) (string, error) {
	byLocale, ok := r.subjectRegistry[templateKey]
	if !ok {
		return "", fmt.Errorf("unknown email template %q", templateKey)
	}
	if subject, ok := byLocale[locale]; ok {
		return subject, nil
	}
	if subject, ok := byLocale[domain.LocaleEnglish]; ok {
		return subject, nil
	}
	return "", fmt.Errorf("missing subject template for %q", templateKey)
}

func (r *TemplateRenderer) renderHTML(templateKey string, locale domain.Locale, ctx map[string]any) (string, error) {
	localeDir := r.localeDir(templateKey, locale, ".html")
	basePath := "templates/" + localeDir + "/base.html"
	bodyPath := "templates/" + localeDir + "/" + templateKey + ".html"
	tpl, err := htmltemplate.New("base").ParseFS(r.fs, basePath, bodyPath)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := tpl.ExecuteTemplate(&b, "base.html", ctx); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (r *TemplateRenderer) renderText(templateKey string, locale domain.Locale, ctx map[string]any) (string, error) {
	localeDir := r.localeDir(templateKey, locale, ".txt")
	bodyPath := "templates/" + localeDir + "/" + templateKey + ".txt"
	content, err := fs.ReadFile(r.fs, bodyPath)
	if err != nil {
		return "", err
	}
	return renderTextString(string(content), ctx)
}

func (r *TemplateRenderer) localeDir(templateKey string, locale domain.Locale, ext string) string {
	loc := domain.NormalizeLocale(string(locale))
	path := "templates/" + string(loc) + "/" + templateKey + ext
	if _, err := fs.Stat(r.fs, path); err == nil {
		return string(loc)
	}
	return string(domain.LocaleEnglish)
}

func renderTextString(tpl string, ctx map[string]any) (string, error) {
	t, err := texttemplate.New("text").Parse(tpl)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := t.Execute(&b, ctx); err != nil {
		return "", err
	}
	return b.String(), nil
}

func buildTemplateContext(brand BrandContext, data map[string]any) map[string]any {
	ctx := map[string]any{
		"tenant_name":     strings.TrimSpace(brand.TenantName),
		"tenant_logo_url": strings.TrimSpace(brand.TenantLogoURL),
		"primary_url":     strings.TrimSpace(brand.PrimaryURL),
		"support_email":   strings.TrimSpace(brand.SupportEmail),
	}
	if ctx["tenant_name"] == "" {
		ctx["tenant_name"] = "Elevare"
	}
	for k, v := range data {
		ctx[k] = v
	}
	return ctx
}
