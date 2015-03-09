package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

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

type ByDate []BlogEntry

func (a ByDate) Len() int { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Date < a[j].Date }

type BlogEntry struct {
	Title, Date string
	MarkdownFile string
}

func (e BlogEntry)Permalink() string {
	extension := filepath.Ext(e.MarkdownFile)
	htmlFile := filepath.Base(e.MarkdownFile)
	return htmlFile[0:len(htmlFile)-len(extension)] + ".html"
}

func (e *BlogEntry)buildPage() Page {
	title, date, size := getMetaData(e.MarkdownFile)
	e.Title = title
	e.Date = date
	page := Page {
		Article: e,
		Content: getContent(e.MarkdownFile, int64(size)),
		Permalink: e.Permalink(),
	}
	return page
}

type Page struct {
	Article *BlogEntry
	Content template.HTML
	Permalink string
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

func getMetaData(filename string) (title, date string, size int) {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var s string
	scanner.Scan()
	s = scanner.Text()
	size = len(s)
	title = strings.TrimPrefix(s, "Title:")
	scanner.Scan()
	s = scanner.Text()
	size += len(s)
	date = strings.TrimPrefix(s, "Date:")
	size += 2 // Account for linefeed
	return
}

func getContent(filename string, off int64) template.HTML {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()
	file.Seek(off, 0)
	fi, _ := file.Stat()
	markdown := make([]byte, fi.Size() - off)
	file.Read(markdown)
	return template.HTML(blackfriday.MarkdownCommon(markdown))
}

func writePage(page *Page, url string, tmpl *template.Template) {
	f, _ := os.Create(url)
	defer f.Close()
	err := tmpl.Execute(f, page)
	check(err)
}

func init() {
	flag.StringVar(&cssFile, "css", "", "CCS file")
	flag.StringVar(&tmplFile, "template", "", "HTML template file")
	flag.StringVar(&aboutFile, "about", "", "Markdown file for about page")
}

func main() {
	flag.Parse()
	articlePath := flag.Arg(0)
	if articlePath == "" {
		os.Exit(1)
	}

	var htmlTemplate *template.Template
	var err error
	if tmplFile != "" {
		htmlTemplate, err = template.ParseFiles(tmplFile)
	} else {
		htmlTemplate, err = template.New("blog").Parse(defaultTemplate)
	}
	check(err)

	var entries []BlogEntry

	// Build and write blog entries pages
	filenames, _ := filepath.Glob(filepath.Join(articlePath, "*.md"))
	for _, filename := range(filenames) {
		entry := BlogEntry { MarkdownFile: filename }
		articlePage := entry.buildPage()
		writePage(&articlePage, articlePage.Permalink, htmlTemplate)
		entries = append(entries, entry)
	}

	// Build archives links and write archives page
	linkTemplate, _ := template.New("links").Parse("<p><a href={{.Permalink}}>{{.Title}}</a></p>")
	check(err)
	var buf bytes.Buffer
	sort.Sort(ByDate(entries))
	for _, entry := range(entries) {
		linkTemplate.Execute(&buf, entry)
	}
	archivePage := Page { Content: template.HTML(buf.String()) }
	writePage(&archivePage, "archives.html", htmlTemplate)

	// Get newest entry and build index page
	indexPage := entries[len(entries)-1].buildPage()
	writePage(&indexPage, "index.html", htmlTemplate)

	if aboutFile != "" {
		aboutPage := Page { Content: getContent(aboutFile, 0) }
		writePage(&aboutPage, "about.html", htmlTemplate)
	}
}
