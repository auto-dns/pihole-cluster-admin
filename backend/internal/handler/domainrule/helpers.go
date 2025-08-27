package domainrule

import (
	"fmt"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
)

func parseRuleType(s string) (domain.RuleType, bool) {
	switch s {
	case "allow":
		return domain.RuleTypeAllow, true
	case "deny":
		return domain.RuleTypeDeny, true
	default:
		return "", false
	}
}
func parseRuleKind(s string) (domain.RuleKind, bool) {
	switch s {
	case "exact":
		return domain.RuleKindExact, true
	case "regex":
		return domain.RuleKindRegex, true
	default:
		return "", false
	}
}
func normalizeDomains(v any) ([]string, error) {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil, fmt.Errorf("domain must not be empty")
		}
		return []string{t}, nil
	case []any:
		out := make([]string, 0, len(t))
		for _, it := range t {
			s, ok := it.(string)
			if !ok {
				return nil, fmt.Errorf("domain list must contain only strings")
			}
			if s == "" {
				return nil, fmt.Errorf("domain must not be empty")
			}
			out = append(out, s)
		}
		if len(out) == 0 {
			return nil, fmt.Errorf("domain list must not be empty")
		}
		return out, nil
	default:
		return nil, fmt.Errorf("domain must be string or array of strings")
	}
}
