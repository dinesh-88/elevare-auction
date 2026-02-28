package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/dinesh-88/elevare-auction/internal/domain"
	"github.com/dinesh-88/elevare-auction/internal/port"
	"github.com/google/uuid"
)

type OnboardingUseCase struct {
	Tenants port.TenantRepository
	Fees    port.FeeRepository
	Taxes   port.TaxRepository
	Hasher  port.PasswordHasher
}

type InstitutionOnboardingData struct {
	InstitutionName      string  `json:"institution_name"`
	Country              string  `json:"country"`
	LogoURL              string  `json:"logo_url"`
	PrimaryColor         string  `json:"primary_color"`
	SecondaryColor       string  `json:"secondary_color"`
	FaviconURL           string  `json:"favicon_url"`
	FontFamily           string  `json:"font_family"`
	DefaultCurrency      string  `json:"default_currency"`
	VATRatePercent       float64 `json:"vat_rate_percent"`
	BuyersPremiumPercent float64 `json:"buyers_premium_percent"`
	LegalTermsMarkdown   string  `json:"legal_terms_markdown"`
	SupportEmail         string  `json:"support_email"`
}

type InstitutionOnboardingState struct {
	Status      domain.OnboardingStatus   `json:"status"`
	Step        int                       `json:"step"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
	Data        InstitutionOnboardingData `json:"data"`
}

type SaveInstitutionOnboardingInput struct {
	Step int                       `json:"step"`
	Data InstitutionOnboardingData `json:"data"`
}

type TenantSignupInput struct {
	OrganizationName       string
	DesiredSubdomain       string
	AdminEmail             string
	Password               string
	Country                string
	EstimatedAuctionVolume *int
}

var subdomainPattern = regexp.MustCompile(`^[a-z0-9-]{3,40}$`)

func normalizeSubdomain(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	s = strings.Split(s, ".")[0]
	return s
}

func (uc *OnboardingUseCase) PublicSignup(ctx context.Context, in TenantSignupInput) (*domain.Tenant, *domain.User, error) {
	name := strings.TrimSpace(in.OrganizationName)
	subdomain := normalizeSubdomain(in.DesiredSubdomain)
	email := strings.ToLower(strings.TrimSpace(in.AdminEmail))
	country := strings.TrimSpace(in.Country)

	if name == "" || email == "" || country == "" || len(in.Password) < 8 || !subdomainPattern.MatchString(subdomain) {
		return nil, nil, ErrValidation
	}

	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return nil, nil, err
	}

	tenant := &domain.Tenant{
		ID:                     uuid.New(),
		Name:                   name,
		Subdomain:              subdomain,
		Status:                 domain.TenantPending,
		Country:                country,
		EstimatedAuctionVolume: in.EstimatedAuctionVolume,
		RequireAdminApproval:   true,
		RequireKYC:             false,
	}
	admin := &domain.User{
		ID:             uuid.New(),
		TenantID:       tenant.ID,
		Email:          email,
		PasswordHash:   hash,
		Role:           domain.RoleOwner,
		Status:         domain.UserStatusActive,
		EmailVerified:  false,
		ApprovalStatus: domain.ApprovalPending,
		KYCStatus:      domain.KYCNone,
	}

	err = uc.Tenants.CreatePendingTenantWithAdmin(ctx, port.CreatePendingTenantWithAdminInput{
		TenantID:               tenant.ID,
		TenantName:             tenant.Name,
		Subdomain:              tenant.Subdomain,
		Country:                tenant.Country,
		EstimatedAuctionVolume: tenant.EstimatedAuctionVolume,
		RequireAdminApproval:   tenant.RequireAdminApproval,
		RequireKYC:             tenant.RequireKYC,
		AdminUserID:            admin.ID,
		AdminEmail:             admin.Email,
		AdminPasswordHash:      admin.PasswordHash,
		AdminEmailVerified:     admin.EmailVerified,
		AdminApprovalStatus:    admin.ApprovalStatus,
	})
	if err != nil {
		return nil, nil, err
	}

	return tenant, admin, nil
}

func (uc *OnboardingUseCase) ListTenantRequests(ctx context.Context, status domain.TenantStatus) ([]domain.Tenant, error) {
	if status == "" {
		status = domain.TenantPending
	}
	return uc.Tenants.ListByStatus(ctx, status)
}

func (uc *OnboardingUseCase) ApproveTenant(ctx context.Context, tenantID uuid.UUID) error {
	return uc.Tenants.UpdateStatusAndAdminApproval(ctx, tenantID, domain.TenantActive, domain.ApprovalApproved)
}

func (uc *OnboardingUseCase) RejectTenant(ctx context.Context, tenantID uuid.UUID) error {
	return uc.Tenants.UpdateStatusAndAdminApproval(ctx, tenantID, domain.TenantRejected, domain.ApprovalRejected)
}

func (uc *OnboardingUseCase) GetInstitutionOnboarding(ctx context.Context, tenantID uuid.UUID) (*InstitutionOnboardingState, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return uc.buildInstitutionState(ctx, tenant)
}

func (uc *OnboardingUseCase) SaveInstitutionOnboarding(ctx context.Context, tenantID uuid.UUID, in SaveInstitutionOnboardingInput) (*InstitutionOnboardingState, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	if in.Step < 1 || in.Step > 6 {
		return nil, ErrValidation
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	data := sanitizeOnboardingData(in.Data)

	switch in.Step {
	case 1:
		if data.InstitutionName == "" || data.Country == "" {
			return nil, ErrValidation
		}
		if err := uc.Tenants.UpdateProfile(ctx, tenantID, data.InstitutionName, data.Country); err != nil {
			return nil, err
		}
	case 2:
		if err := uc.saveSiteState(ctx, tenant, data); err != nil {
			return nil, err
		}
	case 3:
		if len(data.DefaultCurrency) != 3 {
			return nil, ErrValidation
		}
		if err := uc.Tenants.UpdateDefaultCurrency(ctx, tenantID, data.DefaultCurrency); err != nil {
			return nil, err
		}
		if err := uc.ensureDefaultVAT(ctx, tenantID, data.VATRatePercent); err != nil {
			return nil, err
		}
	case 4:
		if data.BuyersPremiumPercent < 0 {
			return nil, ErrValidation
		}
		if err := uc.ensureDefaultBuyerPremium(ctx, tenantID, data.BuyersPremiumPercent); err != nil {
			return nil, err
		}
	case 5:
		if strings.TrimSpace(data.LegalTermsMarkdown) == "" {
			return nil, ErrValidation
		}
		if err := uc.saveSiteState(ctx, tenant, data); err != nil {
			return nil, err
		}
	case 6:
		if !strings.Contains(data.SupportEmail, "@") {
			return nil, ErrValidation
		}
		if err := uc.saveSiteState(ctx, tenant, data); err != nil {
			return nil, err
		}
	}

	nextStatus := tenant.OnboardingStatus
	if nextStatus == "" || nextStatus == domain.OnboardingNotStarted {
		nextStatus = domain.OnboardingInProgress
	}
	nextStep := tenant.OnboardingStep
	if in.Step > nextStep {
		nextStep = in.Step
	}
	if err := uc.Tenants.UpdateOnboardingState(ctx, tenantID, nextStatus, nextStep, tenant.OnboardingCompletedAt); err != nil {
		return nil, err
	}
	updated, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return uc.buildInstitutionState(ctx, updated)
}

func (uc *OnboardingUseCase) CompleteInstitutionOnboarding(ctx context.Context, tenantID uuid.UUID) (*InstitutionOnboardingState, error) {
	if err := ensureTenantScope(ctx, tenantID); err != nil {
		return nil, err
	}
	tenant, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	state, err := uc.buildInstitutionState(ctx, tenant)
	if err != nil {
		return nil, err
	}
	if err := validateInstitutionCompletion(state.Data); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := uc.Tenants.UpdateOnboardingState(ctx, tenantID, domain.OnboardingCompleted, 7, &now); err != nil {
		return nil, err
	}
	updated, err := uc.Tenants.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return uc.buildInstitutionState(ctx, updated)
}

func sanitizeOnboardingData(in InstitutionOnboardingData) InstitutionOnboardingData {
	in.InstitutionName = strings.TrimSpace(in.InstitutionName)
	in.Country = strings.TrimSpace(in.Country)
	in.LogoURL = strings.TrimSpace(in.LogoURL)
	in.PrimaryColor = strings.TrimSpace(in.PrimaryColor)
	in.SecondaryColor = strings.TrimSpace(in.SecondaryColor)
	in.FaviconURL = strings.TrimSpace(in.FaviconURL)
	in.FontFamily = strings.TrimSpace(in.FontFamily)
	in.DefaultCurrency = strings.ToUpper(strings.TrimSpace(in.DefaultCurrency))
	in.LegalTermsMarkdown = strings.TrimSpace(in.LegalTermsMarkdown)
	in.SupportEmail = strings.TrimSpace(in.SupportEmail)
	in.VATRatePercent = roundRate(in.VATRatePercent)
	in.BuyersPremiumPercent = roundRate(in.BuyersPremiumPercent)
	return in
}

func roundRate(value float64) float64 {
	return math.Round(value*100) / 100
}

func (uc *OnboardingUseCase) buildInstitutionState(ctx context.Context, tenant *domain.Tenant) (*InstitutionOnboardingState, error) {
	state := &InstitutionOnboardingState{
		Status:      tenant.OnboardingStatus,
		Step:        tenant.OnboardingStep,
		CompletedAt: tenant.OnboardingCompletedAt,
		Data: InstitutionOnboardingData{
			InstitutionName: strings.TrimSpace(tenant.Name),
			Country:         strings.TrimSpace(tenant.Country),
			DefaultCurrency: strings.ToUpper(strings.TrimSpace(tenant.DefaultCurrency)),
		},
	}
	if state.Status == "" {
		state.Status = domain.OnboardingNotStarted
	}
	branding := decodeJSONMap(tenant.BrandingDraft)
	if len(branding) == 0 {
		branding = decodeJSONMap(tenant.Branding)
	}
	content := decodeJSONMap(tenant.ContentDraft)
	if len(content) == 0 {
		content = decodeJSONMap(tenant.ContentPublished)
	}

	state.Data.LogoURL = mapString(branding, "logo_url")
	state.Data.PrimaryColor = mapString(branding, "primary_color")
	state.Data.SecondaryColor = mapString(branding, "secondary_color")
	state.Data.FaviconURL = mapString(branding, "favicon_url")
	state.Data.FontFamily = mapString(branding, "font_family")
	state.Data.LegalTermsMarkdown = nestedStringMap(content, "legal", "terms_markdown")
	state.Data.SupportEmail = nestedStringMap(content, "contact", "email")
	if state.Data.SupportEmail == "" {
		state.Data.SupportEmail = nestedStringMap(content, "footer", "contact_email")
	}

	if uc.Fees != nil {
		if value, err := uc.currentBuyerPremium(ctx, tenant.ID); err == nil {
			state.Data.BuyersPremiumPercent = value
		}
	}
	if uc.Taxes != nil {
		if value, err := uc.currentVATRate(ctx, tenant.ID); err == nil {
			state.Data.VATRatePercent = value
		}
	}
	return state, nil
}

func (uc *OnboardingUseCase) currentVATRate(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	rules, err := uc.Taxes.ListRules(ctx, tenantID, false)
	if err != nil {
		return 0, err
	}
	for _, rule := range rules {
		if strings.EqualFold(strings.TrimSpace(rule.Name), "VAT") {
			return rule.RatePercent, nil
		}
	}
	return 0, nil
}

func (uc *OnboardingUseCase) currentBuyerPremium(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	components, err := uc.Fees.ListComponents(ctx, tenantID, false)
	if err != nil {
		return 0, err
	}
	for _, component := range components {
		if strings.EqualFold(strings.TrimSpace(component.Name), "Buyer's Premium") {
			return component.Value, nil
		}
	}
	return 0, nil
}

func (uc *OnboardingUseCase) ensureDefaultVAT(ctx context.Context, tenantID uuid.UUID, rate float64) error {
	if uc.Taxes == nil {
		return nil
	}
	rules, err := uc.Taxes.ListRules(ctx, tenantID, false)
	if err != nil {
		return err
	}
	var target *domain.TaxRule
	for i := range rules {
		if strings.EqualFold(strings.TrimSpace(rules[i].Name), "VAT") {
			target = &rules[i]
			break
		}
	}
	if target == nil {
		target = &domain.TaxRule{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "VAT",
			RatePercent: rate,
			Active:      true,
		}
		if err := uc.Taxes.CreateRule(ctx, target); err != nil {
			return err
		}
	} else {
		target.Name = "VAT"
		target.RatePercent = rate
		target.Active = true
		if err := uc.Taxes.UpdateRule(ctx, target); err != nil {
			return err
		}
	}
	assignments, err := uc.Taxes.ListAssignments(ctx, tenantID, domain.TaxRuleScopeTenant, tenantID)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		if assignment.TaxRuleID == target.ID {
			return nil
		}
	}
	return uc.Taxes.CreateAssignment(ctx, tenantID, &domain.TaxRuleAssignment{
		ID:        uuid.New(),
		ScopeType: domain.TaxRuleScopeTenant,
		ScopeID:   tenantID,
		TaxRuleID: target.ID,
	})
}

func (uc *OnboardingUseCase) ensureDefaultBuyerPremium(ctx context.Context, tenantID uuid.UUID, rate float64) error {
	if uc.Fees == nil {
		return nil
	}
	components, err := uc.Fees.ListComponents(ctx, tenantID, false)
	if err != nil {
		return err
	}
	var target *domain.FeeComponent
	for i := range components {
		if strings.EqualFold(strings.TrimSpace(components[i].Name), "Buyer's Premium") {
			target = &components[i]
			break
		}
	}
	if target == nil {
		target = &domain.FeeComponent{
			ID:              uuid.New(),
			TenantID:        tenantID,
			Name:            "Buyer's Premium",
			CalculationType: domain.FeeCalculationPercentage,
			Value:           rate,
			AppliesTo:       domain.FeeAppliesToHammer,
			Taxable:         true,
			Active:          true,
		}
		if err := uc.Fees.CreateComponent(ctx, target); err != nil {
			return err
		}
	} else {
		target.Name = "Buyer's Premium"
		target.CalculationType = domain.FeeCalculationPercentage
		target.Value = rate
		target.AppliesTo = domain.FeeAppliesToHammer
		target.Taxable = true
		target.Active = true
		if err := uc.Fees.UpdateComponent(ctx, target); err != nil {
			return err
		}
	}
	assignments, err := uc.Fees.ListAssignments(ctx, tenantID, domain.FeeScopeTenant, tenantID)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		if assignment.FeeComponentID == target.ID {
			return nil
		}
	}
	return uc.Fees.CreateAssignment(ctx, tenantID, &domain.FeeAssignment{
		ID:             uuid.New(),
		ScopeType:      domain.FeeScopeTenant,
		ScopeID:        tenantID,
		FeeComponentID: target.ID,
	})
}

func (uc *OnboardingUseCase) saveSiteState(ctx context.Context, tenant *domain.Tenant, data InstitutionOnboardingData) error {
	branding := decodeJSONMap(tenant.BrandingDraft)
	if len(branding) == 0 {
		branding = decodeJSONMap(tenant.Branding)
	}
	content := decodeJSONMap(tenant.ContentDraft)
	if len(content) == 0 {
		content = decodeJSONMap(tenant.ContentPublished)
	}

	setMapString(branding, data.LogoURL, "logo_url")
	setMapString(branding, data.PrimaryColor, "primary_color")
	setMapString(branding, data.SecondaryColor, "secondary_color")
	setMapString(branding, data.FaviconURL, "favicon_url")
	setMapString(branding, data.FontFamily, "font_family")

	if data.LegalTermsMarkdown != "" {
		setNestedString(content, data.LegalTermsMarkdown, "legal", "terms_markdown")
	}
	if data.SupportEmail != "" {
		setNestedString(content, data.SupportEmail, "contact", "email")
		setNestedString(content, data.SupportEmail, "footer", "contact_email")
	}

	brandingRaw, err := json.Marshal(branding)
	if err != nil {
		return err
	}
	contentRaw, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return uc.Tenants.UpdateSiteDraft(ctx, tenant.ID, brandingRaw, contentRaw)
}

func validateInstitutionCompletion(data InstitutionOnboardingData) error {
	if strings.TrimSpace(data.InstitutionName) == "" || strings.TrimSpace(data.Country) == "" {
		return fmt.Errorf("complete identity before activation")
	}
	if len(strings.TrimSpace(data.DefaultCurrency)) != 3 {
		return fmt.Errorf("set a default currency before activation")
	}
	if data.BuyersPremiumPercent < 0 {
		return fmt.Errorf("set valid fee defaults before activation")
	}
	if strings.TrimSpace(data.LegalTermsMarkdown) == "" {
		return fmt.Errorf("add legal terms before activation")
	}
	if !strings.Contains(strings.TrimSpace(data.SupportEmail), "@") {
		return fmt.Errorf("configure a support email before activation")
	}
	return nil
}

func decodeJSONMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func mapString(input map[string]any, key string) string {
	value, _ := input[key].(string)
	return strings.TrimSpace(value)
}

func nestedStringMap(input map[string]any, path ...string) string {
	current := input
	for i, part := range path {
		value, ok := current[part]
		if !ok {
			return ""
		}
		if i == len(path)-1 {
			text, _ := value.(string)
			return strings.TrimSpace(text)
		}
		next, ok := value.(map[string]any)
		if !ok {
			return ""
		}
		current = next
	}
	return ""
}

func setMapString(input map[string]any, value string, key string) {
	if input == nil {
		return
	}
	if strings.TrimSpace(value) == "" {
		delete(input, key)
		return
	}
	input[key] = strings.TrimSpace(value)
}

func setNestedString(input map[string]any, value string, path ...string) {
	if input == nil || len(path) == 0 {
		return
	}
	current := input
	for _, part := range path[:len(path)-1] {
		next, ok := current[part].(map[string]any)
		if !ok || next == nil {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
	last := path[len(path)-1]
	if strings.TrimSpace(value) == "" {
		delete(current, last)
		return
	}
	current[last] = strings.TrimSpace(value)
}
