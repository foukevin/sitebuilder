package main

import (
	"bufio"
	"bytes"
	"flag"
	//"fmt"
	"html/template"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var cssFile string
var tmplFile string
var aboutFile string
var blogName string
var blogURL string
var authorName string
var authorEmail string

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
func (a ByDate) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

type BlogEntry struct {
	Title string
	Date time.Time
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
		return p.Article.Date.Format("2006-01-02")
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

func getMetaData(filename string) (title string, date time.Time, size int) {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var s string
	scanner.Scan()
	s = scanner.Text()
	size = len(s)
	title = strings.TrimPrefix(s, "Title: ")
	scanner.Scan()
	s = scanner.Text()
	size += len(s)
	s = strings.TrimPrefix(s, "Date: ")
	date, _ = time.Parse("2006-01-02", s)
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
	flag.StringVar(&cssFile, "css", "", "CSS file")
	flag.StringVar(&tmplFile, "template", "", "HTML template file")
	flag.StringVar(&aboutFile, "about", "", "Markdown file for about page")

	user, err := user.Current()
	check(err)
	flag.StringVar(&blogName, "name", "", "Blog name")
	flag.StringVar(&blogURL, "url", "", "Blog URL")
	flag.StringVar(&authorName, "author", user.Name, "Author name")
	flag.StringVar(&authorEmail, "email", "", "Author email")
}

func writeFeed(entries []BlogEntry) {
	author := &feeds.Author{ authorName, authorEmail }
	feed := &feeds.Feed{
		Title:       "blog",
		Link:        &feeds.Link{Href: "/"},
		Author:      author,
		Created:     time.Now(),
	}

	for _, entry := range(entries) {
		link := blogURL + "/" + entry.Permalink()
		item := &feeds.Item {
			Title: entry.Title,
			Link: &feeds.Link{ Href: link },
			Author: author,
			Created: entry.Date,
		}
		feed.Items = append(feed.Items, item)
	}

	atom, _ := feed.ToAtom()
	rss, _ := feed.ToRss()

	file, err := os.Create("atom.xml")
	check(err)
	defer file.Close()

	file.WriteString(atom)
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

	writeFeed(entries)
}
