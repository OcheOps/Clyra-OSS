// Package pricing is the single source of truth for plan prices, limits,
// and marketing-facing feature lists. It supports geo-aware responses (NGN/USD).
package pricing

// Plan identifier.
type Plan string

const (
	PlanFree Plan = "free"
	PlanPro  Plan = "pro"
	PlanBYO  Plan = "byo"
)

// Region codes. "" means default / unknown (treated as global).
const (
	RegionNigeria = "NG"
	RegionGlobal  = ""
)

// PlanCard is the front-end shape for a single plan entry.
type PlanCard struct {
	Plan         Plan     `json:"plan"`
	Name         string   `json:"name"`
	Tagline      string   `json:"tagline"`
	AmountMinor  int      `json:"amount_minor"` // smallest unit of currency
	AmountMajor  string   `json:"amount_major"` // display-ready
	Currency     string   `json:"currency"`     // "NGN" | "USD" | ""
	CurrencySym  string   `json:"currency_symbol"`
	Period       string   `json:"period"` // "month"
	DailyLimit   int      `json:"daily_limit"`
	Features     []string `json:"features"`
	CTA          string   `json:"cta"`
	Recommended  bool     `json:"recommended,omitempty"`
	RequiresKeys bool     `json:"requires_keys,omitempty"`
}

// Catalog is the localized pricing + plan metadata returned to the client.
type Catalog struct {
	Region  string     `json:"region"`
	Country string     `json:"country,omitempty"`
	Plans   []PlanCard `json:"plans"`
}

// Build returns the catalog for a given region. Region codes are ISO-3166-1 alpha-2.
// Anything outside known regions falls back to the global (USD) catalog.
func Build(region string) Catalog {
	switch region {
	case RegionNigeria:
		return Catalog{
			Region:  RegionNigeria,
			Country: "Nigeria",
			Plans: []PlanCard{
				{
					Plan: PlanFree, Name: "Free", Tagline: "Get started, no card",
					AmountMinor: 0, AmountMajor: "₦0", Currency: "NGN", CurrencySym: "₦", Period: "month",
					DailyLimit: 10, CTA: "Start free",
					Features: []string{
						"10 AI generations per account",
						"Solution, Network & Enterprise modes",
						"1 Terraform / CloudFormation export",
						"2,000-character prompts",
						"Local-only — projects stay on this device",
					},
				},
				{
					Plan: PlanPro, Name: "Pro", Tagline: "Managed AI, no setup",
					AmountMinor: 350000, AmountMajor: "₦3,500", Currency: "NGN", CurrencySym: "₦", Period: "month",
					DailyLimit: 200, CTA: "Upgrade to Pro", Recommended: true,
					Features: []string{
						"200 AI generations per account",
						"Managed Bedrock — no API keys needed",
						"Priority model routing",
						"Unlimited IaC exports",
						"8,000-character prompts",
						"Cloud sync across devices",
						"Everything in Free",
					},
				},
				{
					Plan: PlanBYO, Name: "BYO", Tagline: "Bring your own keys",
					AmountMinor: 200000, AmountMajor: "₦2,000", Currency: "NGN", CurrencySym: "₦", Period: "month",
					DailyLimit: 1000, CTA: "Go BYO", RequiresKeys: true,
					Features: []string{
						"1,000 AI generations per account",
						"Use your OpenAI / Anthropic / Gemini keys",
						"Keys stored in your browser only",
						"8,000-character prompts",
						"Cloud sync across devices",
						"Everything in Free",
					},
				},
			},
		}
	default:
		return Catalog{
			Region:  RegionGlobal,
			Country: "",
			Plans: []PlanCard{
				{
					Plan: PlanFree, Name: "Free", Tagline: "Get started, no card",
					AmountMinor: 0, AmountMajor: "$0", Currency: "USD", CurrencySym: "$", Period: "month",
					DailyLimit: 10, CTA: "Start free",
					Features: []string{
						"10 AI generations per account",
						"Solution, Network & Enterprise modes",
						"1 Terraform / CloudFormation export",
						"2,000-character prompts",
						"Local-only — projects stay on this device",
					},
				},
				{
					Plan: PlanPro, Name: "Pro", Tagline: "Managed AI, no setup",
					AmountMinor: 1000, AmountMajor: "$10", Currency: "USD", CurrencySym: "$", Period: "month",
					DailyLimit: 200, CTA: "Upgrade to Pro", Recommended: true,
					Features: []string{
						"200 AI generations per account",
						"Managed Bedrock — no API keys needed",
						"Priority model routing",
						"Unlimited IaC exports",
						"8,000-character prompts",
						"Cloud sync across devices",
						"Everything in Free",
					},
				},
				{
					Plan: PlanBYO, Name: "BYO", Tagline: "Bring your own keys",
					AmountMinor: 400, AmountMajor: "$4", Currency: "USD", CurrencySym: "$", Period: "month",
					DailyLimit: 1000, CTA: "Go BYO", RequiresKeys: true,
					Features: []string{
						"1,000 AI generations per account",
						"Use your OpenAI / Anthropic / Gemini keys",
						"Keys stored in your browser only",
						"8,000-character prompts",
						"Cloud sync across devices",
						"Everything in Free",
					},
				},
			},
		}
	}
}

// CardFor returns the price/card metadata for a single (region, plan) pair.
// Returns (PlanCard{}, false) if the plan isn't in the region's catalog.
func CardFor(region string, plan Plan) (PlanCard, bool) {
	cat := Build(region)
	for _, p := range cat.Plans {
		if p.Plan == plan {
			return p, true
		}
	}
	return PlanCard{}, false
}
