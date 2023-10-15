/*
	Spindle
	A static site generator
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import "os"
import "fmt"
import "strings"
import "unicode/utf8"

import "github.com/mattn/go-isatty"

var running_in_term = false

func init() {
	running_in_term = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func print(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
}

func println(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
	os.Stdout.WriteString("\n")
}

func printf(format string, guff ...any) {
	fmt.Fprintf(os.Stdout, format, guff...)
}

func eprint(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stderr.WriteString(w)
		if i < l {
			os.Stderr.WriteString(" ")
		}
	}
}

func eprintln(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stderr.WriteString(w)
		if i < l {
			os.Stderr.WriteString(" ")
		}
	}
	os.Stderr.WriteString("\n")
}

func eprintf(format string, guff ...any) {
	fmt.Fprintf(os.Stderr, format, guff...)
	os.Stderr.WriteString("\n")
}

func apply_color(input string) string {
	const ansi_reset = "\033[0m"
	const ansi_color = "\033[91m"

	buffer := strings.Builder{}
	buffer.Grow(len(input) + 128)

	last_rune := 'x'

	for {
		if len(input) == 0 {
			break
		}

		r, w := utf8.DecodeRuneInString(input)
		input = input[w:]

		if r == '$' {
			last_rune = r
			continue
		}

		if last_rune == '$' {
			last_rune = 'x'

			if r == '0' || r == '1' {
				if !running_in_term {
					continue
				} else if r == '0' {
					buffer.WriteString(ansi_reset)
				} else {
					buffer.WriteString(ansi_color)
				}
			} else {
				buffer.WriteRune('$')
				buffer.WriteRune(r)
			}

			continue
		}

		last_rune = r
		buffer.WriteRune(r)
	}

	return buffer.String()
}