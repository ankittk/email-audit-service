package rulesengine

import (
	"strings"
	"sync"

	pb "github.com/ankittk/email-audit-service/proto"
)

type RulesEngine struct {
	mu    sync.RWMutex
	rules map[string]map[string]*pb.Rule
}

func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		rules: make(map[string]map[string]*pb.Rule),
	}
}

func (re *RulesEngine) AddRule(companyID string, rule *pb.Rule) string {
	re.mu.Lock()
	defer re.mu.Unlock()

	if _, ok := re.rules[companyID]; !ok {
		re.rules[companyID] = make(map[string]*pb.Rule)
	}
	re.rules[companyID][rule.Id] = rule
	return rule.Id
}

func (re *RulesEngine) UpdateRule(companyID string, rule *pb.Rule) bool {
	re.mu.Lock()
	defer re.mu.Unlock()

	companyRules, ok := re.rules[companyID]
	if !ok {
		return false
	}

	if _, exists := companyRules[rule.Id]; !exists {
		return false
	}

	companyRules[rule.Id] = rule
	return true
}

func (re *RulesEngine) DeleteRule(companyID, ruleID string) bool {
	re.mu.Lock()
	defer re.mu.Unlock()

	companyRules, ok := re.rules[companyID]
	if !ok {
		return false
	}

	if _, exists := companyRules[ruleID]; !exists {
		return false
	}

	delete(companyRules, ruleID)
	return true
}

func (re *RulesEngine) ListRules(companyID, category string, enabledOnly bool) []*pb.Rule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	var list []*pb.Rule
	companyRules, ok := re.rules[companyID]
	if !ok {
		return list
	}

	for _, r := range companyRules {
		if enabledOnly && !r.Enabled {
			continue
		}
		if category != "" && !strings.EqualFold(r.Category, category) {
			continue
		}
		list = append(list, r)
	}
	return list
}

func (re *RulesEngine) EvaluateRules(ruleSet []*pb.Rule, thread *pb.EmailThread) ([]*pb.RuleEvaluation, float64) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	var (
		totalWeight   float64
		weightedScore float64
		evaluations   []*pb.RuleEvaluation
	)

	for _, rule := range ruleSet {
		if !rule.Enabled {
			continue
		}

		var passed bool
		var suggestions []string

		switch rule.Type {
		case pb.RuleType_REGEX_BASED:
			passed, suggestions = evaluateRegexRule(rule, thread)
		case pb.RuleType_KEYWORD_BASED:
			passed, suggestions = evaluateKeywordRule(rule, thread)
		default:
			passed = false
			suggestions = []string{"Unsupported rule type"}
		}

		score := 0.0
		if passed {
			score = 1.0
		}

		weightedScore += score * rule.Weight
		totalWeight += rule.Weight

		eval := &pb.RuleEvaluation{
			RuleId:      rule.Id,
			RuleName:    rule.Name,
			Category:    rule.Category,
			Weight:      rule.Weight,
			Passed:      passed,
			Score:       score,
			Suggestions: suggestions,
		}
		evaluations = append(evaluations, eval)
	}

	var overallScore float64
	if totalWeight > 0 {
		overallScore = (weightedScore / totalWeight) * 100
	}

	return evaluations, overallScore
}
