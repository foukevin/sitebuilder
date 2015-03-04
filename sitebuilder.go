package main

import (
	"os"
	"io/ioutil"
	"html/template"
	"github.com/russross/blackfriday"
)

type Page struct {
	Title string
	Body template.HTML
}

func loadPage(filename string) (*Page, error) {
	markdown, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	body := template.HTML(blackfriday.MarkdownBasic(markdown))
	return &Page{Title: filename, Body: body}, nil
}

func main() {
	page, _ := loadPage("articles/1.md")
	tmpl, _ := template.ParseFiles("blog.html")
	tmpl.Execute(os.Stdout, page)
}
