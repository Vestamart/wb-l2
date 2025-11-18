package unpack

import (
	"errors"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	runes := []rune(s)
	var result []rune

	var prev rune
	hasPrev := false
	escape := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escape {
			result = append(result, r)
			prev = r
			hasPrev = true
			escape = false
			continue
		}

		if r == '\\' {
			escape = true
			continue
		}

		if unicode.IsDigit(r) {
			if !hasPrev {
				return "", ErrInvalidString
			}

			j := i
			for j < len(runes) && unicode.IsDigit(runes[j]) {
				j++
			}

			count := 0
			for k := i; k < j; k++ {
				count = count*10 + int(runes[k]-'0')
			}
			i = j - 1

			if count == 0 {
				if len(result) == 0 {
					return "", ErrInvalidString
				}
				result = result[:len(result)-1]
				continue
			}

			for k := 0; k < count-1; k++ {
				result = append(result, prev)
			}
			continue
		}

		result = append(result, r)
		prev = r
		hasPrev = true
	}

	if escape {
		return "", ErrInvalidString
	}

	return string(result), nil
}
