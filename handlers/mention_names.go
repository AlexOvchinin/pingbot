package handlers

import (
	"sort"
	"strings"
)

// parseMentionNames preserves a single (possibly spaced) name, accepts comma
// separated names, and recognizes existing multi-word names in a space-separated list.
func parseMentionNames(payload string, existing []string) []string {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil
	}
	if strings.Contains(payload, ",") {
		return uniqueNames(strings.Split(payload, ","))
	}
	known := append([]string(nil), existing...)
	sort.Slice(known, func(i, j int) bool { return len(known[i]) > len(known[j]) })
	var names []string
	remainder := payload
	matched := false
	for remainder != "" {
		found := false
		for _, name := range known {
			if strings.HasPrefix(remainder, name) && (len(remainder) == len(name) || remainder[len(name)] == ' ') {
				names = append(names, name)
				remainder = strings.TrimSpace(remainder[len(name):])
				matched, found = true, true
				break
			}
		}
		if !found {
			word, rest, _ := strings.Cut(remainder, " ")
			names = append(names, word)
			remainder = strings.TrimSpace(rest)
		}
	}
	if !matched {
		return []string{payload}
	}
	return uniqueNames(names)
}

func uniqueNames(names []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			result = append(result, name)
			seen[name] = true
		}
	}
	return result
}
