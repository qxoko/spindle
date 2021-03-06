package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type markup struct {
	vars map[string]string
	data []*markup_object

	pos int
}

type markup_object struct {
	object_type uint8
	offset      uint8
	text        []string
	vars        map[string]string
}

const (
	PARAGRAPH uint8 = iota
	RAW_TEXT
	HEADING
	LIST_O
	LIST_U
	MEDIA
	IMAGE
	DIVIDER
	FUNCTION
	FUNCTION_INLINE
	IMPORT
	CHUNK
	WHITESPACE
	BLOCK_CODE
	BLOCK_END
	BLOCK
	BLOCK_ELSE
	BLOCK_IF
)

type parser_def struct {
	enum uint8
	trim bool
	args bool
}

var single_token = map[rune]*parser_def {
	'*': &parser_def {
		enum: RAW_TEXT,
	},
	'#': &parser_def {
		enum: HEADING,
		trim: true,
	},
	'%': &parser_def {
		enum: IMAGE,
		args: true,
	},
	'.': &parser_def {
		enum: PARAGRAPH,
		trim: true,
	},
	'-': &parser_def {
		enum: LIST_U,
		trim: true,
	},
	'+': &parser_def {
		enum: LIST_O,
		trim: true,
	},
	'@': &parser_def {
		enum: MEDIA,
		args: true,
	},
	'>': &parser_def {
		enum: CHUNK,
		args: true,
	},
	'~': &parser_def {
		enum: IMPORT,
		args: true,
	},
	'ø': &parser_def {
		enum: FUNCTION,
		args: true,
	},
	'$': &parser_def {
		enum: FUNCTION_INLINE,
		trim: true,
	},
}

func markup_parser(input string) *markup {
	input = strings.TrimSpace(input)

	data_list := make([]*markup_object, 0, 64)

	vars_stack := make([]map[string]string, 0, 3)
	vars_stack = append(vars_stack, make(map[string]string, 16))

	active_map := vars_stack[0]

	pop := func() {
		if len(vars_stack) == 1 {
			active_map = vars_stack[0]
		}
		vars_stack = vars_stack[:len(vars_stack) - 1]
		active_map = vars_stack[len(vars_stack) - 1]
	}

	push := func() {
		vars_stack = append(vars_stack, make(map[string]string, 16))
		active_map = vars_stack[len(vars_stack) - 1]
	}

	for {
		input = consume_whitespace(input)

		if len(input) == 0 {
			break
		}

		if input[0] == '/' && len(input) > 1 && input[1] == '/' {
			line := extract_to_newline(input)
			input = input[len(line):]

			the_rune, _ := utf8.DecodeLastRuneInString(line)

			if the_rune == '{' {
				input = input[len(extract_code_block(input))+1:] // +1 trailing brace
			}

			continue
		}

		the_rune, rune_width := utf8.DecodeRuneInString(input)

		rune_count := count_rune(input, the_rune)
		rune_width = rune_count * rune_width

		switch the_rune {
		case '}': // we're closing a block
			pop()
			input = input[1:]
			data_list = append(data_list, &markup_object{
				object_type: BLOCK_END,
			})
			continue

		case '-': // we have a divider (special case)
			if rune_count >= 3 {
				input = input[rune_width:]
				data_list = append(data_list, &markup_object{
					object_type: DIVIDER,
				})
				continue
			}
		}

		if info, ok := single_token[the_rune]; ok {
			input = input[rune_width:]
			data := extract_to_newline(input)
			input = input[len(data):]

			if info.trim {
				data = strings.TrimSpace(data)
			}

			obj := &markup_object{
				object_type: info.enum,
				offset:      uint8(rune_count),
			}

			if info.args {
				obj.text = unix_args(data)
			} else {
				obj.text = []string{data}
			}

			data_list = append(data_list, obj)

			// whitespace check
			if has_double_newline(input) {
				data_list = append(data_list, &markup_object {
					object_type: WHITESPACE,
				})
			}

			continue
		}

		ident := extract_ident(input)

		var_test := consume_whitespace(input[len(ident):])

		if len(var_test) > 0 && var_test[0] == '=' {
			var_test = consume_whitespace(var_test[1:])

			data := ""

			if var_test[0] == '`' {
				for i, c := range var_test[1:] {
					if c == '`' {
						data = var_test[1:i+1]
						var_test = var_test[i+2:]
						break
					}
				}

				if consume_whitespace(data)[0] == '<' {
					data = clean_html(data)
				}

			} else {
				data = extract_to_newline(var_test)
				var_test = var_test[len(data):]
				data = strings.TrimSpace(data)
			}

			switch data {
			case "":       console_print("bad variable in page") // @error
			case "false":  data = "0"
			}

			active_map[ident] = data
			input = var_test
			continue
		}

		line := extract_to_newline(input)
		the_rune, _ = utf8.DecodeLastRuneInString(line)

		// we are a block
		if the_rune == '{' {
			args := strings.TrimSpace(line[len(ident) : len(line)-1])
			args_valid := false

			if args != "" && len(extract_arbitrary_word(args)) == len(args) {
				args_valid = true
			}

			input = input[len(line):]

			push()

			the_token := &markup_object {
				vars: active_map,
			}

			switch ident {
			case "if":
				if !args_valid {
					console_print("unknown guff in 'if' statement") // @error
				}

				the_token.object_type = BLOCK_IF

				if args[0] == '!' {
					the_token.text = []string{args[1:]}
					the_token.offset = 1
				} else {
					the_token.text = []string{args}
				}

				data_list = append(data_list, the_token)
				continue

			case "else":
				the_token.object_type = BLOCK_ELSE
				data_list = append(data_list, the_token)
				continue

			case "code":
				code := extract_code_block(input)

				input = input[len(code)+1:] // +1 trailing brace

				x := make([]string, 0, 2)
				x = append(x, process_code(code))

				if args_valid {
					x = append(x, args)
				}

				the_token.object_type = BLOCK_CODE
				the_token.text = x

				data_list = append(data_list, the_token)
				continue

			case "function":
				program_text := extract_code_block(input)
				input = input[len(program_text)+1:] // +1 trailing brace

				the_token.object_type = FUNCTION_INLINE
				the_token.text = []string{program_text}

				data_list = append(data_list, the_token)
				continue

			case "html":
				html := extract_code_block(input)
				input = input[len(html)+1:] // +1 trailing brace
				html = clean_html(html)

				the_token.object_type = RAW_TEXT
				the_token.text = []string{html}

				data_list = append(data_list, the_token)
				continue
			}

			// block with userland ident
			the_token.object_type = BLOCK
			the_token.text = []string{ident}

			data_list = append(data_list, the_token)
			continue
		}

		// paragraph
		data_list = append(data_list, &markup_object{
			object_type: PARAGRAPH,
			text:        []string{strings.TrimSpace(line)},
		})
		input = input[len(line):]
	}

	return &markup{
		data: data_list,
		vars: vars_stack[0],
	}
}

func extract_code_block(input string) string {
	depth := 1
	lastr := 'a'

	for i, c := range input {
		if lastr == '\\' {
			lastr = c
			continue
		}

		switch c {
		case '{':
			depth++
		case '}':
			depth--
		}

		lastr = c

		if depth == 0 {
			return input[:i]
		}
	}

	return input
}

func assign_plate(some_page *markup) {
	ident, ok := some_page.vars["plate"]

	// no plate, just merge config
	if !ok {
		some_page.vars = merge_vars(some_page.vars, config.vars)
		return
	}

	var the_plate *markup

	if x, ok := cache_plate[ident]; ok {
		the_plate = x
	} else {
		raw_text, ok := load_file(sprint("config/plates/%s.x", ident))

		if !ok {
			fmt.Printf("plate %q does not exist\n", ident)
			return
		}

		the_plate = markup_parser(raw_text)

		assign_plate(the_plate)

		cache_plate[ident] = the_plate
	}

	some_page.vars = merge_vars(some_page.vars, the_plate.vars)
}

func process_vars(vars map[string]string) map[string]string {
	image_prefix, _ := vars["image_prefix"]

	for key, value := range vars {
		if strings.HasSuffix(key, "image") {
			if config.build_mode && is_draft(value) {
				fmt.Printf("image: %q is draft\n", value) // @warning
			}

			vars[key] = rewrite_image_path(value, image_prefix, config.image_rewrite_extensions)
		}
	}

	return vars
}