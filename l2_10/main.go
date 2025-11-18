package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type sortOptions struct {
	keyField int  // -k
	numeric  bool // -n
	reverse  bool // -r
	unique   bool // -u

	month   bool // -M
	trimEnd bool // -b
	check   bool // -c
	human   bool // -h
}

type record struct {
	line     string
	key      string
	num      float64
	numValid bool
	monthVal int
}

func main() {
	opts := parseFlags()

	records, err := readRecords(os.Stdin, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.check {
		if !isSorted(records, opts) {
			fmt.Fprintln(os.Stderr, "data is not sorted")
			os.Exit(1)
		}
		// данные отсортированы, просто выходим
		return
	}

	sortRecords(records, opts)

	if opts.unique {
		writeUnique(os.Stdout, records)
	} else {
		writeAll(os.Stdout, records)
	}
}

func parseFlags() sortOptions {
	keyField := flag.Int("k", 0, "sort by tab-separated column number (1-based)")
	numeric := flag.Bool("n", false, "numeric sort")
	reverse := flag.Bool("r", false, "reverse sort")
	unique := flag.Bool("u", false, "output only unique lines")
	month := flag.Bool("M", false, "sort by month name (Jan, Feb, ...)")
	trimEnd := flag.Bool("b", false, "ignore trailing blanks when comparing")
	check := flag.Bool("c", false, "check whether data is sorted")
	human := flag.Bool("h", false, "human numeric sort (e.g. 10K, 2M)")

	flag.Parse()

	return sortOptions{
		keyField: *keyField,
		numeric:  *numeric,
		reverse:  *reverse,
		unique:   *unique,
		month:    *month,
		trimEnd:  *trimEnd,
		check:    *check,
		human:    *human,
	}
}

func readRecords(f *os.File, opts sortOptions) ([]record, error) {
	scanner := bufio.NewScanner(f)
	// увеличим буфер, чтобы нормально читать длинные строки
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var records []record
	for scanner.Scan() {
		line := scanner.Text()
		rec := makeRecord(line, opts)
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func makeRecord(line string, opts sortOptions) record {
	key := line

	if opts.keyField > 0 {
		cols := strings.Split(line, "\t")
		if opts.keyField <= len(cols) {
			key = cols[opts.keyField-1]
		} else {
			key = ""
		}
	}

	if opts.trimEnd {
		key = strings.TrimRightFunc(key, unicode.IsSpace)
	}

	rec := record{
		line: line,
		key:  key,
	}

	if opts.month {
		rec.monthVal = parseMonth(key)
	}

	if opts.numeric || opts.human {
		num, ok := parseNumber(key, opts.human)
		if ok {
			rec.num = num
			rec.numValid = true
		}
	}

	return rec
}

func sortRecords(records []record, opts sortOptions) {
	if len(records) == 0 {
		return
	}

	sort.SliceStable(records, func(i, j int) bool {
		cmp := compareRecords(records[i], records[j], opts)
		if opts.reverse {
			return cmp > 0
		}
		return cmp < 0
	})
}

func compareRecords(a, b record, opts sortOptions) int {
	// M: сортировка по месяцу
	if opts.month {
		if a.monthVal != b.monthVal {
			if a.monthVal < b.monthVal {
				return -1
			}
			return 1
		}
		// если месяцы равны или не распознаны — падаем дальше к строковому сравнению
	}

	// n/h: числовая сортировка
	if opts.numeric || opts.human {
		if a.numValid && b.numValid && a.num != b.num {
			if a.num < b.num {
				return -1
			}
			return 1
		}
		if a.numValid != b.numValid {
			if a.numValid {
				return -1
			}
			return 1
		}
		// если численные значения равны или оба невалидные — переходим к лексикографическому сравнению
	}

	// строковое сравнение по ключу
	if a.key < b.key {
		return -1
	}
	if a.key > b.key {
		return 1
	}

	// чтобы порядок был детерминированным, добиваемся сравнения всей строки
	if a.line < b.line {
		return -1
	}
	if a.line > b.line {
		return 1
	}

	return 0
}

func isSorted(records []record, opts sortOptions) bool {
	for i := 1; i < len(records); i++ {
		if compareRecords(records[i-1], records[i], opts) > 0 {
			return false
		}
	}
	return true
}

func writeAll(w *os.File, records []record) {
	for _, r := range records {
		fmt.Fprintln(w, r.line)
	}
}

func writeUnique(w *os.File, records []record) {
	if len(records) == 0 {
		return
	}
	prev := records[0].line
	fmt.Fprintln(w, prev)

	for i := 1; i < len(records); i++ {
		if records[i].line != prev {
			fmt.Fprintln(w, records[i].line)
			prev = records[i].line
		}
	}
}

// parseNumber парсит число, при human=true понимает суффиксы K, M, G, T.
func parseNumber(s string, human bool) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	if !human {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}

	return parseHumanNumber(s)
}

func parseHumanNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	runes := []rune(s)
	last := runes[len(runes)-1]

	multiplier := 1.0
	switch unicode.ToUpper(last) {
	case 'K':
		multiplier = 1024
	case 'M':
		multiplier = 1024 * 1024
	case 'G':
		multiplier = 1024 * 1024 * 1024
	case 'T':
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		// нет суффикса — обычное число
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}

	numStr := string(runes[:len(runes)-1])
	v, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, false
	}

	return v * multiplier, true
}

func parseMonth(s string) int {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return 0
	}
	name := strings.ToLower(s[:3])

	switch name {
	case "jan":
		return 1
	case "feb":
		return 2
	case "mar":
		return 3
	case "apr":
		return 4
	case "may":
		return 5
	case "jun":
		return 6
	case "jul":
		return 7
	case "aug":
		return 8
	case "sep":
		return 9
	case "oct":
		return 10
	case "nov":
		return 11
	case "dec":
		return 12
	default:
		return 0
	}
}
