package main

import (
	"regexp"
	// "strings"
)

/*
	To be clear, I hate this.

	Unfortunately, there isn't a _great_ way of letting users
	define their own inline syntaxes and then have the
	parser figure those out — it would require that the
	parser inform the parser of how the parser should work
	mid-stream which... is possible, of course, but it's so
	far beyond my capabilities as a programmer that I
	haven't managed to even begin to reason about how to
	architect something that could work that way.

	I did make a (never checked-in) earnest attempt at *italics*
	and **bold**-style definitions, which are based on
	balanced sets of known characters, and therefore quite
	easy to specify (pseudo-syntax, I never did work out an
	idiomatic style):

		[**]:begin = <b>
		[**]:end   = </b>

	Unfortunately, this falls down immediately because it's not
	sufficient for specifying something like this:

		1. [some link](https://website.com)
		2. some mapping syntax?
		3. <a href="%2">%1<a>

	To do this, I'd need a secondary definition layer (step 2)
	and this is where my point in the opening statement
	comes from: the parser would have to be made of aware of
	the how the stream is assembled _while parsing_ and
	context-switch in and out of it by scope.

	Rather than letting this kill the project, I've decided to
	just ship _something_ that works, and as distasteful as
	this is amid the true parser that handles everything else,
	I really only have the choice of either hard-coding
	inline syntaxes or letting users configure them
	external to the program text.
*/

type Regex_Config struct {
	Input  string `toml:"pattern"`
	Output string `toml:"template"`
}

type regex_entry struct {
	regexp *regexp.Regexp
	output []byte
	// d_hash uint32
}

func process_regex_array(array []*Regex_Config) ([]*regex_entry, bool) {
	output := make([]*regex_entry, 0, len(array))

	for _, entry := range array {
		re, err := regexp.Compile(entry.Input)
		if err != nil {
			return nil, false
		}

		output = append(output, &regex_entry{
			regexp: re,
			output: []byte(entry.Output),
			// d_hash: new_hash(entry.Output),
		})
	}

	return output, true
}

func apply_regex_array(array []*regex_entry, input string) string {
	if len(array) == 0 {
		return input
	}

	cast := []byte(input)

	for _, entry := range array {
		cast = entry.regexp.ReplaceAll(cast, entry.output)
	}

	return string(cast)
}

/*func _apply_regex_array(r *renderer, array []*regex_entry, input string) string {
	if len(array) == 0 {
		return input
	}

	// array := []*regexp.Regexp{
	// 	regexp.MustCompile(`\[(.+?)\]\((.+?)\)`),
	// 	regexp.MustCompile(`\*(\S(.+?)\S)\*`),
	// }

	for _, entry := range array {
		wrapper_block, ok := r.get_in_scope(entry.d_hash)
		if ok {
			did_push := r.push_blank_scope(immediate_decl_count(wrapper_block.get_children()))

			// indexes := entry.FindAllStringSubmatchIndex(edit, -1)
			match_groups := entry.FindAllStringSubmatch(edit, -1)

			r.render_ast(spindle, page, p.content)

			// @todo replace this with index call to make it _go fasta_
			for _, match_group := range match_groups {
				edit = strings.ReplaceAll(edit, match_group[0], match_group[1])
			}

			if did_push { r.pop_scope() }
		}
	}

	fmt.Println(edit)
}*/