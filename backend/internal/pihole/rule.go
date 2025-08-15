package pihole

import "strings"

type RuleType string

const (
	RuleTypeAllow RuleType = "allow"
	RuleTypeDeny  RuleType = "deny"
)

func ParseRuleType(s string) (RuleType, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(RuleTypeAllow):
		return RuleTypeAllow, true
	case string(RuleTypeDeny):
		return RuleTypeDeny, true
	default:
		return "", false
	}
}

type RuleKind string

const (
	RuleKindExact RuleKind = "exact"
	RuleKindRegex RuleKind = "regex"
)

func ParseRuleKind(s string) (RuleKind, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(RuleKindExact):
		return RuleKindExact, true
	case string(RuleKindRegex):
		return RuleKindRegex, true
	default:
		return "", false
	}
}
