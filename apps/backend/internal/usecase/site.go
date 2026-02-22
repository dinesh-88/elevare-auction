package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/dinesh-88/elevare-auction/internal/authz"
	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type SiteUseCase struct {
	Tenants port.TenantRepository
}

type SiteDraftInput struct {
	Branding json.RawMessage
	Content  json.RawMessage
}

func (uc *SiteUseCase) Get(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if err := authz.Authorize(ctx, domain.PermissionSiteEdit); err != nil {
		return nil, ErrForbidden
	}
	return uc.Tenants.GetByID(ctx, tenantID)
}

func (uc *SiteUseCase) UpdateDraft(ctx context.Context, tenantID uuid.UUID, in SiteDraftInput) error {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return err
	}
	if err := authz.Authorize(ctx, domain.PermissionSiteEdit); err != nil {
		return ErrForbidden
	}
	branding, err := normalizeBranding(in.Branding)
	if err != nil {
		return err
	}
	content, err := normalizeContent(in.Content)
	if err != nil {
		return err
	}
	return uc.Tenants.UpdateSiteDraft(ctx, tenantID, branding, content)
}

func (uc *SiteUseCase) Publish(ctx context.Context, tenantID uuid.UUID) (int, int, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return 0, 0, err
	}
	if err := authz.Authorize(ctx, domain.PermissionSitePublish); err != nil {
		return 0, 0, ErrForbidden
	}
	return uc.Tenants.PublishSiteDraft(ctx, tenantID)
}

type brandingInput struct {
	LogoURL        string `json:"logo_url"`
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	FaviconURL     string `json:"favicon_url"`
	FontFamily     string `json:"font_family"`
}

type siteContent struct {
	Homepage struct {
		Hero struct {
			Headline       string `json:"headline"`
			Subheadline    string `json:"subheadline"`
			BannerImageURL string `json:"banner_image_url"`
			CTA            struct {
				Label string `json:"label"`
				Href  string `json:"href"`
			} `json:"cta"`
		} `json:"hero"`
		FeaturedAuction struct {
			Mode      string     `json:"mode"`
			AuctionID *uuid.UUID `json:"auction_id,omitempty"`
		} `json:"featuredAuction"`
		Highlights []struct {
			Title string `json:"title"`
			Body  string `json:"body"`
			Icon  string `json:"icon"`
		} `json:"highlights"`
	} `json:"homepage"`
	Footer struct {
		Text         string            `json:"text"`
		ContactEmail string            `json:"contact_email"`
		SocialLinks  map[string]string `json:"social_links,omitempty"`
	} `json:"footer"`
	About struct {
		Markdown string `json:"markdown"`
	} `json:"about"`
	Contact struct {
		Address string `json:"address"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
	} `json:"contact"`
}

var colorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{6})$`)
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func normalizeBranding(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return []byte(`{}`), nil
	}
	var in brandingInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, ErrValidation
	}
	in.LogoURL = strings.TrimSpace(in.LogoURL)
	in.PrimaryColor = strings.TrimSpace(in.PrimaryColor)
	in.SecondaryColor = strings.TrimSpace(in.SecondaryColor)
	in.FaviconURL = strings.TrimSpace(in.FaviconURL)
	in.FontFamily = strings.TrimSpace(in.FontFamily)

	if in.PrimaryColor != "" && !colorRegex.MatchString(in.PrimaryColor) {
		return nil, errors.New("invalid primary_color")
	}
	if in.SecondaryColor != "" && !colorRegex.MatchString(in.SecondaryColor) {
		return nil, errors.New("invalid secondary_color")
	}
	if err := validateURLOrInternal(in.LogoURL); err != nil {
		return nil, errors.New("invalid logo_url")
	}
	if err := validateURLOrInternal(in.FaviconURL); err != nil {
		return nil, errors.New("invalid favicon_url")
	}
	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func normalizeContent(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return []byte(`{}`), nil
	}
	var in siteContent
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, ErrValidation
	}
	in.Homepage.Hero.Headline = trimTo(in.Homepage.Hero.Headline, 80)
	in.Homepage.Hero.Subheadline = trimTo(in.Homepage.Hero.Subheadline, 160)
	in.Homepage.Hero.CTA.Label = trimTo(in.Homepage.Hero.CTA.Label, 40)
	if err := validateURLOrInternal(strings.TrimSpace(in.Homepage.Hero.BannerImageURL)); err != nil {
		return nil, errors.New("invalid homepage.hero.banner_image_url")
	}
	if err := validateURLOrInternal(strings.TrimSpace(in.Homepage.Hero.CTA.Href)); err != nil {
		return nil, errors.New("invalid homepage.hero.cta.href")
	}

	mode := strings.ToUpper(strings.TrimSpace(in.Homepage.FeaturedAuction.Mode))
	if mode == "" {
		mode = "AUTO"
	}
	if mode != "AUTO" && mode != "MANUAL" {
		return nil, errors.New("invalid homepage.featuredAuction.mode")
	}
	in.Homepage.FeaturedAuction.Mode = mode
	if mode == "MANUAL" && in.Homepage.FeaturedAuction.AuctionID == nil {
		return nil, errors.New("manual featured auction requires auction_id")
	}
	if len(in.Homepage.Highlights) > 6 {
		in.Homepage.Highlights = in.Homepage.Highlights[:6]
	}
	for i := range in.Homepage.Highlights {
		in.Homepage.Highlights[i].Title = trimTo(in.Homepage.Highlights[i].Title, 60)
		in.Homepage.Highlights[i].Body = trimTo(in.Homepage.Highlights[i].Body, 160)
		in.Homepage.Highlights[i].Icon = trimTo(in.Homepage.Highlights[i].Icon, 30)
	}
	in.Footer.Text = trimTo(in.Footer.Text, 160)
	in.Footer.ContactEmail = strings.TrimSpace(in.Footer.ContactEmail)
	if in.Footer.ContactEmail != "" && !strings.Contains(in.Footer.ContactEmail, "@") {
		return nil, errors.New("invalid footer.contact_email")
	}
	for k, v := range in.Footer.SocialLinks {
		if err := validateURLOrInternal(v); err != nil {
			return nil, errors.New("invalid footer.social_links." + k)
		}
		in.Footer.SocialLinks[k] = strings.TrimSpace(v)
	}

	in.About.Markdown = sanitizeMarkdown(in.About.Markdown)
	in.Contact.Address = trimTo(in.Contact.Address, 250)
	in.Contact.Phone = trimTo(in.Contact.Phone, 60)
	in.Contact.Email = strings.TrimSpace(in.Contact.Email)
	if in.Contact.Email != "" && !strings.Contains(in.Contact.Email, "@") {
		return nil, errors.New("invalid contact.email")
	}

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func trimTo(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func validateURLOrInternal(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(raw, "/") {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http(s) links allowed")
	}
	if strings.EqualFold(u.Scheme, "javascript") {
		return errors.New("script urls are forbidden")
	}
	return nil
}

func sanitizeMarkdown(s string) string {
	s = strings.TrimSpace(s)
	s = htmlTagRegex.ReplaceAllString(s, "")
	return s
}
