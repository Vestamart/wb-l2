package main

import (
	"fmt"
	"github.com/Vestamart/wb-l2/tree/main/l2_11/anagram"
)

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}

	groups := anagram.FindAnagrams(words)

	for key, vals := range groups {
		fmt.Printf("%q: %v\n", key, vals)
	}
}
