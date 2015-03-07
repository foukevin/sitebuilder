package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type Page struct {
	Title, Date, Permalink, Latest string
	IsArticle bool
	Content template.HTML
}

type BlogEntry struct {
	Title, Date string
	Permalink string
	MarkdownFile string
}

func PageName(filename string) string {
	var extension = filepath.Ext(filename)
	var htmlFile = filepath.Base(filename)
	return htmlFile[0:len(htmlFile)-len(extension)] + ".html"
}

func LoadBlogEntries(filename string) []BlogEntry {
	csvfile, err := os.Open(filename)
	defer csvfile.Close()
	if err != nil {
		panic(err)
	}

	reader := csv.NewReader(csvfile)
	reader.FieldsPerRecord = -1
        rawCSVdata, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var entries []BlogEntry
	for _, each := range rawCSVdata[1:] {
		markdownFile := "articles/" + each[2]
		entry := BlogEntry{
			Date: each[0],
			Title: each[1],
			MarkdownFile: markdownFile,
			Permalink: PageName(markdownFile),
		}
		entries = append(entries, entry)
	}

	return entries
}

func GetContentFromMarkdown(filename string) template.HTML {
	markdown, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return template.HTML(blackfriday.MarkdownCommon(markdown))
}

func BuildPage(page *Page, tmpl *template.Template) {
	f, _ := os.Create(page.Permalink)
	defer f.Close()

	err := tmpl.Execute(f, page)
	if err!= nil {
		panic(err)
	}
}

type ByDate []BlogEntry

func (a ByDate) Len() int { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Date < a[j].Date }

func main() {
	flag.Parse()
	filename := flag.Arg(0)

	htmlTemplate, err := template.ParseFiles("templates/blog.html")
	if err != nil {
		panic(err)
	}
	linkTemplate, _ := template.New("links").Parse("<p><a href={{.Permalink}}>{{.Title}}</a></p>")
	if err != nil {
		panic(err)
	}

	entries := LoadBlogEntries(filename)
	sort.Sort(ByDate(entries))
	latest := entries[len(entries)-1].Permalink

	var buf bytes.Buffer
	for _, entry := range(entries) {
		articlePage := Page {
			Title: entry.Title, Date: entry.Date, Latest: latest,
			Permalink: entry.Permalink, IsArticle: true,
			Content: GetContentFromMarkdown(entry.MarkdownFile),
		}
		BuildPage(&articlePage, htmlTemplate)
		linkTemplate.Execute(&buf, entry)
	}

	archivePage := Page {
		Title: "Archive", Latest: latest, Permalink: "archives.html",
		IsArticle: false, Content: template.HTML(buf.String()),
	}
	BuildPage(&archivePage, htmlTemplate)

	aboutPage := Page {
		Title: "About me", Latest: latest, Permalink: "about.html",
		IsArticle: false, Content: GetContentFromMarkdown("articles/about.md"),
	}
	BuildPage(&aboutPage, htmlTemplate)
}
