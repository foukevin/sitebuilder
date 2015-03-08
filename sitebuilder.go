package main

import (
	"bufio"
	//"fmt"
	"bytes"
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
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

type BlogEntry struct {
	Title, Date string
	MarkdownFile string
}

func (e BlogEntry)Permalink() string {
	extension := filepath.Ext(e.MarkdownFile)
	htmlFile := filepath.Base(e.MarkdownFile)
	return htmlFile[0:len(htmlFile)-len(extension)] + ".html"
}

type Page struct {
	Article *BlogEntry
	Content template.HTML
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

func getContentFromMarkdown(filename string) template.HTML {
	markdown, err := ioutil.ReadFile(filename)
	check(err)
	return template.HTML(blackfriday.MarkdownCommon(markdown))
}

func buildPage(page *Page, url string, tmpl *template.Template) {
	f, _ := os.Create(url)
	defer f.Close()

	err := tmpl.Execute(f, page)
	check(err)
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

func getMetaData(filename string) (string, string) {
	file, err := os.Open(filename)
		check(err)
		defer file.Close()

		scanner := bufio.NewScanner(file)

		var s string
		scanner.Scan()
		s = scanner.Text()
		title := strings.TrimPrefix(s, "Title:")
		scanner.Scan()
		s = scanner.Text()
		date := strings.TrimPrefix(s, "Date:")
		return title, date
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

	linkTemplate, _ := template.New("links").Parse("<p><a href={{.Permalink}}>{{.Title}}</a></p>")
	check(err)

	var entries []BlogEntry

	filenames, _ := filepath.Glob(filepath.Join(articlePath, "*.md"))

	for _, filename := range(filenames) {
		title, date := getMetaData(filename)
		entry := BlogEntry{
			Date: date,
			Title: title,
			MarkdownFile: filename,
		}

		markdown, err := ioutil.ReadFile(filename)
		content := template.HTML(blackfriday.MarkdownCommon(markdown))
		check(err)
		articlePage := Page {
			Article: &entry,
			Content: content,
		}
		buildPage(&articlePage, articlePage.Permalink(), htmlTemplate)

		entries = append(entries, entry)
	}

	var buf bytes.Buffer
	sort.Sort(ByDate(entries))
	for _, entry := range(entries) {
		linkTemplate.Execute(&buf, entry)
	}

	indexPage := Page { Article: &entries[len(entries)-1] }
	buildPage(&indexPage, "index.html", htmlTemplate)

	archivePage := Page { Content: template.HTML(buf.String()) }
	buildPage(&archivePage, "archives.html", htmlTemplate)

	if aboutFile != "" {
		aboutPage := Page { Content: getContentFromMarkdown(aboutFile) }
		buildPage(&aboutPage, "about.html", htmlTemplate)
	}
}
