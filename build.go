package main

import (
	"sync"
	"sort"
	"strings"
	"path/filepath"
)

func build_project(args []string) {
	public_dir := "public"

	if len(args) > 0 {
		public_dir = args[0]
	}

	init_minify() // for copy_mini

	files, folders := get_files("source", public_dir, true)

	make_directory(public_dir)

	for _, folder := range folders {
		make_directory(folder.output)
	}

	// currently non-thread safe
	for _, file := range files {
		if file.file_type == MARKUP {
			page_obj, _ := load_page(file.source, true)
			make_file(file.output, markup_render(page_obj))
		}
	}

	sitemap(files, public_dir)

	var wg sync.WaitGroup

	for _, pointer_file := range files {
		file := *pointer_file

		switch file.file_type {
		case MARKUP:
			continue

		case STATIC_JS, STATIC_CSS:
			go func() {
				wg.Add(1)
				defer wg.Done()

				copy_mini(&file)
			}()

		case IMAGE_JPG, IMAGE_PNG:
			go func() {
				wg.Add(1)
				defer wg.Done()

				image_handler(&file, config.image_target)
			}()

		default:
			go func() {
				wg.Add(1)
				defer wg.Done()

				copy_file(file.source, file.output)
			}()
		}
	}

	wg.Wait()

	console_handler.flush()
}

func load_page(path string, no_drafts bool) (*markup, bool) {
	raw_text, ok := load_file(path)

	if !ok {
		return nil, false
	}

	page_obj := markup_parser(raw_text)

	page_obj.no_drafts = no_drafts

	assign_plate(page_obj)

	page_obj.vars = process_vars(page_obj, page_obj.vars)

	page_obj.vars["raw_path"]   = path
	page_obj.vars["url_pretty"] = make_url_from_path(path[6:])
	// @todo unpretty url?

	return page_obj, true
}

func make_url_from_path(input string) string {
	input = input[:len(input) - len(filepath.Ext(input))]

	if strings.HasSuffix(input, "index") {
		input = input[:len(input) - 5]
	}

	return filepath.ToSlash(filepath.Clean(input))
}

func sitemap(files []*file, public_dir string) {
	ordered := make([]string, len(files) + 24)

	for _, file := range files {
		switch file.file_type {
		case MARKUP, STATIC_HTML:
			path := make_url_from_path(file.source[6:])
			ordered = append(ordered, sprint(sitemap_entry, join_url(config.vars["domain"], path)))
		default:
			continue
		}
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i] < ordered[j]
	})

	buffer := strings.Builder {}
	buffer.Grow(len(sitemap_template) + len(ordered) * len(sitemap_entry))

	buffer.WriteString(sitemap_template)

	for _, page := range ordered {
		buffer.WriteString(page)
	}

	buffer.WriteString(`</urlset>`)

	make_file(filepath.Join(public_dir, "sitemap.xml"), buffer.String())
}