package rulesengine

import (
	"fmt"
	"regexp"
	"strings"

	pb "github.com/ankittk/email-audit-service/proto"
)

func evaluateRegexRule(rule *pb.Rule, thread *pb.EmailThread) (bool, []string) {
	pattern := rule.Config.GetRegexPattern()
	if pattern == "" {
		return false, []string{"Missing regex pattern in rule config"}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, []string{fmt.Sprintf("Invalid regex pattern: %v", err)}
	}

	for _, msg := range thread.Messages {
		if !re.MatchString(msg.Body) {
			return false, []string{fmt.Sprintf("Message [%s] does not match pattern", msg.MessageId)}
		}
	}
	return true, nil
}

func evaluateKeywordRule(rule *pb.Rule, thread *pb.EmailThread) (bool, []string) {
	keywords := rule.Config.GetKeywords()
	if len(keywords) == 0 {
		return false, []string{"No keywords defined for keyword-based rule"}
	}

	bodyText := strings.ToLower(combineAllBodies(thread))
	missing := []string{}

	for _, keyword := range keywords {
		if !strings.Contains(bodyText, strings.ToLower(keyword)) {
			missing = append(missing, keyword)
		}
	}

	if len(missing) > 0 {
		return false, []string{fmt.Sprintf("Missing keywords: %s", strings.Join(missing, ", "))}
	}

	return true, nil
}

func combineAllBodies(thread *pb.EmailThread) string {
	var builder strings.Builder
	for _, msg := range thread.Messages {
		builder.WriteString(msg.Body)
		builder.WriteString(" ")
	}
	return builder.String()
}

func EvaluateRulesStateless(rules []*pb.Rule, thread *pb.EmailThread) []*pb.RuleEvaluation {
	var evaluations []*pb.RuleEvaluation

	for _, rule := range rules {
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

	return evaluations
}
