package rules

import (
	"fmt"
	"sort"
	"strings"

	"example.com/eventledger/model"
)

type Predicate interface {
	Match(model.EventRecord) bool
	Describe() string
}

type PredicateFunc struct {
	Name string
	Test func(model.EventRecord) bool
}

func (predicate PredicateFunc) Match(value model.EventRecord) bool {
	return predicate.Test != nil && predicate.Test(value)
}

func (predicate PredicateFunc) Describe() string { return predicate.Name }

type All []Predicate

func (predicates All) Match(value model.EventRecord) bool {
	for _, predicate := range predicates {
		if predicate == nil || !predicate.Match(value) {
			return false
		}
	}
	return true
}

func (predicates All) Describe() string {
	parts := make([]string, 0, len(predicates))
	for _, predicate := range predicates {
		if predicate != nil {
			parts = append(parts, predicate.Describe())
		}
	}
	return strings.Join(parts, " and ")
}

type Any []Predicate

func (predicates Any) Match(value model.EventRecord) bool {
	for _, predicate := range predicates {
		if predicate != nil && predicate.Match(value) {
			return true
		}
	}
	return false
}

func (predicates Any) Describe() string {
	parts := make([]string, 0, len(predicates))
	for _, predicate := range predicates {
		if predicate != nil {
			parts = append(parts, predicate.Describe())
		}
	}
	return strings.Join(parts, " or ")
}

type Not struct{ Predicate Predicate }

func (predicate Not) Match(value model.EventRecord) bool {
	return predicate.Predicate != nil && !predicate.Predicate.Match(value)
}

func (predicate Not) Describe() string {
	if predicate.Predicate == nil {
		return "not <nil>"
	}
	return "not " + predicate.Predicate.Describe()
}

func Active() Predicate {
	return PredicateFunc{Name: "active", Test: func(value model.EventRecord) bool { return value.Active }}
}

func Status(status string) Predicate {
	want := strings.ToLower(strings.TrimSpace(status))
	return PredicateFunc{Name: "status=" + want, Test: func(value model.EventRecord) bool {
		return strings.ToLower(strings.TrimSpace(value.Status)) == want
	}}
}

func Tagged(tag string) Predicate {
	want := strings.ToLower(strings.TrimSpace(tag))
	return PredicateFunc{Name: "tag=" + want, Test: func(value model.EventRecord) bool { return value.HasTag(want) }}
}

func MinimumAmount(minimum int64) Predicate {
	return PredicateFunc{Name: fmt.Sprintf("amount>=%d", minimum), Test: func(value model.EventRecord) bool {
		return value.Amount >= minimum
	}}
}

func PriorityBetween(minimum, maximum int) Predicate {
	return PredicateFunc{Name: fmt.Sprintf("priority=%d..%d", minimum, maximum), Test: func(value model.EventRecord) bool {
		return value.Priority >= minimum && value.Priority <= maximum
	}}
}

type Rule struct {
	Name      string
	Predicate Predicate
	Labels    []string
}

func (rule Rule) Validate() error {
	if strings.TrimSpace(rule.Name) == "" {
		return fmt.Errorf("rule name is empty")
	}
	if rule.Predicate == nil {
		return fmt.Errorf("rule %q has no predicate", rule.Name)
	}
	return nil
}

type Set struct{ rules []Rule }

func New(input []Rule) (*Set, error) {
	rules := make([]Rule, len(input))
	copy(rules, input)
	seen := map[string]struct{}{}
	for index := range rules {
		if err := rules[index].Validate(); err != nil {
			return nil, err
		}
		key := strings.ToLower(strings.TrimSpace(rules[index].Name))
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate rule %q", rules[index].Name)
		}
		seen[key] = struct{}{}
		rules[index].Labels = model.NormalizeTags(rules[index].Labels)
	}
	return &Set{rules: rules}, nil
}

func (set *Set) Names() []string {
	if set == nil {
		return []string{}
	}
	result := make([]string, len(set.rules))
	for index, rule := range set.rules {
		result[index] = rule.Name
	}
	sort.Strings(result)
	return result
}

type Evaluation struct {
	Matched []string
	Missed  []string
}

func (set *Set) Evaluate(value model.EventRecord) Evaluation {
	result := Evaluation{Matched: []string{}, Missed: []string{}}
	if set == nil {
		return result
	}
	for _, rule := range set.rules {
		if rule.Predicate.Match(value) {
			result.Matched = append(result.Matched, rule.Name)
		} else {
			result.Missed = append(result.Missed, rule.Name)
		}
	}
	return result
}

func (set *Set) Select(values []model.EventRecord, name string) ([]model.EventRecord, error) {
	if set == nil {
		return nil, fmt.Errorf("rule set is nil")
	}
	for _, rule := range set.rules {
		if rule.Name != name {
			continue
		}
		result := make([]model.EventRecord, 0)
		for _, value := range values {
			if rule.Predicate.Match(value) {
				result = append(result, value.Clone())
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("rule %q not found", name)
}
