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

var cssFile string
var tmplFile string
var aboutFile string
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
{{with .CSS}}<link rel="stylesheet" type="text/css" href={{.}}/>{{end}}
</head>
<body>
<div>
<a href="index.html">home</a>
<a href="archives.html">archives</a>
{{if .HasAboutPage}}<a href="about.html">about</a>{{end}}
<div>
{{with .Title}}<h1>{{.}}</h1>{{end}}
{{if .IsArticle}}<p>{{.Date}}, <a href={{.Permalink}}>permalink</a></p>{{end}}
{{.Content}}
</div>
</body>
</html>`

type BlogEntry struct {
	Title, Date, MarkdownFile string
}

func (e BlogEntry)Permalink() string {
	extension := filepath.Ext(e.MarkdownFile)
	htmlFile := filepath.Base(e.MarkdownFile)
	return htmlFile[0:len(htmlFile)-len(extension)] + ".html"
}

type Page struct {
	Article *BlogEntry
	HTMLContent template.HTML
}

func (p Page)Content() template.HTML {
	if p.Article != nil {
		return GetContentFromMarkdown(p.Article.MarkdownFile)
	}
	return p.HTMLContent
}

func (p Page)Date() string {
	if p.Article != nil {
		return p.Article.Date
	}
	return ""
}

func (p Page)Title() string {
	if p.Article != nil {
		return p.Article.Title
	}
	return ""
}

func (p Page)IsArticle() bool {
	return p.Article != nil
}

func (p Page)HasAboutPage() bool {
	return aboutFile != ""
}

func (p Page)CSS() string {
	return cssFile
}

func (p Page)Permalink() string {
	if (p.Article != nil) {
		return p.Article.Permalink()
	}
	return ""
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
		}
		entries = append(entries, entry)
	}

	sort.Sort(ByDate(entries))
	return entries
}

func GetContentFromMarkdown(filename string) template.HTML {
	markdown, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return template.HTML(blackfriday.MarkdownCommon(markdown))
}

func BuildPage(page *Page, url string, tmpl *template.Template) {
	f, _ := os.Create(url)
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

func init() {
	flag.StringVar(&cssFile, "css", "", "CCS file")
	flag.StringVar(&tmplFile, "template", "", "HTML template file")
	flag.StringVar(&aboutFile, "about", "", "Markdown file for about page")
}

func main() {
	flag.Parse()
	filename := flag.Arg(0)
	if filename == "" {
		os.Exit(1)
	}

	var htmlTemplate *template.Template
	var err error
	if tmplFile != "" {
		htmlTemplate, err = template.ParseFiles(tmplFile)
	} else {
		htmlTemplate, err = template.New("blog").Parse(defaultTemplate)
	}
	if err != nil {
		panic(err)
	}

	linkTemplate, _ := template.New("links").Parse("<p><a href={{.Permalink}}>{{.Title}}</a></p>")
	if err != nil {
		panic(err)
	}

	entries := LoadBlogEntries(filename)
	var buf bytes.Buffer
	for _, entry := range(entries) {
		articlePage := Page { Article: &entry }
		BuildPage(&articlePage, articlePage.Permalink(), htmlTemplate)
		linkTemplate.Execute(&buf, entry)
	}

	indexPage := Page { Article: &entries[len(entries)-1] }
	BuildPage(&indexPage, "index.html", htmlTemplate)

	archivePage := Page { HTMLContent: template.HTML(buf.String()) }
	BuildPage(&archivePage, "archives.html", htmlTemplate)

	if aboutFile != "" {
		aboutPage := Page { HTMLContent: GetContentFromMarkdown(aboutFile) }
		BuildPage(&aboutPage, "about.html", htmlTemplate)
	}
}
