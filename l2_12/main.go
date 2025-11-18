package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type options struct {
	after      int  // -A
	before     int  // -B
	context    int  // -C
	countOnly  bool // -c
	ignoreCase bool // -i
	invert     bool // -v
	fixed      bool // -F
	number     bool // -n
}

func main() {
	opts, pattern, filename := parseFlags()

	matcher, err := buildMatcher(pattern, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var f *os.File
	if filename == "" || filename == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	lines, matches, err := readAndMatchLines(f, matcher, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.countOnly {
		count := 0
		for _, m := range matches {
			if m {
				count++
			}
		}
		fmt.Println(count)
		return
	}

	before := opts.before
	after := opts.after
	if opts.context > 0 {
		before = opts.context
		after = opts.context
	}

	toPrint := make([]bool, len(lines))
	for i, m := range matches {
		if !m {
			continue
		}
		start := i - before
		if start < 0 {
			start = 0
		}
		end := i + after
		if end >= len(lines) {
			end = len(lines) - 1
		}
		for j := start; j <= end; j++ {
			toPrint[j] = true
		}
	}

	hasContext := before > 0 || after > 0
	lastPrinted := -1

	for i := 0; i < len(lines); i++ {
		if !toPrint[i] {
			continue
		}

		if hasContext && lastPrinted != -1 && i > lastPrinted+1 {
			fmt.Println("--")
		}

		if opts.number {
			fmt.Printf("%d:%s\n", i+1, lines[i])
		} else {
			fmt.Println(lines[i])
		}
		lastPrinted = i
	}
}

func parseFlags() (options, string, string) {
	var opts options

	flag.IntVar(&opts.after, "A", 0, "print N lines of trailing context")
	flag.IntVar(&opts.before, "B", 0, "print N lines of leading context")
	flag.IntVar(&opts.context, "C", 0, "print N lines of output context")
	flag.BoolVar(&opts.countOnly, "c", false, "print only a count of matching lines")
	flag.BoolVar(&opts.ignoreCase, "i", false, "ignore case distinctions")
	flag.BoolVar(&opts.invert, "v", false, "invert the sense of matching")
	flag.BoolVar(&opts.fixed, "F", false, "interpret pattern as a fixed string")
	flag.BoolVar(&opts.number, "n", false, "print line number with output lines")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: grep [options] pattern [file]")
		os.Exit(1)
	}

	pattern := args[0]
	filename := ""
	if len(args) > 1 {
		filename = args[1]
	}

	return opts, pattern, filename
}

func buildMatcher(pattern string, opts options) (func(string) bool, error) {
	if opts.fixed {
		if opts.ignoreCase {
			p := strings.ToLower(pattern)
			return func(line string) bool {
				match := strings.Contains(strings.ToLower(line), p)
				if opts.invert {
					return !match
				}
				return match
			}, nil
		}

		return func(line string) bool {
			match := strings.Contains(line, pattern)
			if opts.invert {
				return !match
			}
			return match
		}, nil
	}

	if opts.ignoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return func(line string) bool {
		match := re.MatchString(line)
		if opts.invert {
			return !match
		}
		return match
	}, nil
}

func readAndMatchLines(f *os.File, matcher func(string) bool, opts options) ([]string, []bool, error) {
	scanner := bufio.NewScanner(f)
	// увеличиваем буфер для длинных строк
	const maxCapacity = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	var lines []string
	var matches []bool

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		matches = append(matches, matcher(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return lines, matches, nil
}
