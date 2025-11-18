package anagram

import (
	"sort"
	"strings"
	"unicode/utf8"
)

func FindAnagrams(words []string) map[string][]string {
	type group struct {
		representative string
		words          map[string]struct{}
	}

	signatureToGroup := make(map[string]*group)

	for _, w := range words {
		if w == "" {
			continue
		}

		lower := strings.ToLower(w)
		if utf8.RuneCountInString(lower) == 0 {
			continue
		}

		sig := makeSignature(lower)

		g, ok := signatureToGroup[sig]
		if !ok {
			g = &group{
				representative: lower,
				words:          make(map[string]struct{}),
			}
			signatureToGroup[sig] = g
		}

		g.words[lower] = struct{}{}
	}

	result := make(map[string][]string)

	for _, g := range signatureToGroup {
		if len(g.words) < 2 {
			continue
		}

		list := make([]string, 0, len(g.words))
		for w := range g.words {
			list = append(list, w)
		}
		sort.Strings(list)

		result[g.representative] = list
	}

	return result
}

func makeSignature(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}
