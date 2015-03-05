package main

import (
	"flag"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

var htmlTemplate string

func init() {
	flag.StringVar(&htmlTemplate, "template", "", "HTML template")
}

func main() {
	flag.Parse()
	filename := flag.Arg(0)

	var extension = filepath.Ext(filename)
	var htmlFile = filepath.Base(filename)
	htmlFile = htmlFile[0:len(htmlFile)-len(extension)] + ".html"

	page, err := loadPage(filename)
	if err != nil {
		log.Fatal("Bad file", filename)
	}
	f, _ := os.Create(htmlFile)
	tmpl, _ := template.ParseFiles(htmlTemplate)
	err = tmpl.Execute(f, page)
	if err!= nil {
		log.Fatal("Bad template", htmlTemplate)
	}
}
