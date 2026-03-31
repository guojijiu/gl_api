package parse

import (
	"fmt"
	"regexp"
	"strings"
)

func ExtractTaskUUIDsFromQuestion(question string) []string {
	re := regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	raw := re.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, v := range raw {
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func ExtractNumericIDsFromQuestion(question string) []int {
	re := regexp.MustCompile(`\b\d{3,}\b`)
	raw := re.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[int]struct{}{}
	var out []int
	for _, v := range raw {
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		if err != nil || n <= 0 {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}
