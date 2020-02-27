package main

import (
	"regexp"
	"io/ioutil"
	"net/http"
	"log"
	"html/template"
	"strings"
	"path"
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile("./data/" + filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile("./data/" + filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(templates *template.Template) func(http.ResponseWriter, string, *Page) {
	return func(w http.ResponseWriter, tmpl string, p *Page) {
		err := templates.ExecuteTemplate(w, tmpl + ".html", p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func viewHandler(templates *template.Template) func(http.ResponseWriter, *http.Request, string) {
	return func(w http.ResponseWriter, r *http.Request, title string) {
		p, err := loadPage(title)
		// fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
		if err != nil {
			http.Redirect(w, r, "/edit/" + title, http.StatusFound)
			return
		}
		renderTemplate(templates)(w, "view", p)
	}

}
func editHandler(templates *template.Template) func(http.ResponseWriter, *http.Request, string) {
	return func(w http.ResponseWriter, r *http.Request, title string) {
		p, err := loadPage(title)
		if err != nil {
			p = &Page{Title: title}
		}
		renderTemplate(templates)(w, "edit", p)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(validPath *regexp.Regexp, fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func makeRootHandler(templates *template.Template) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := ioutil.ReadDir("./data")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var titles []string
		for _, file := range files {
			name := file.Name()
			titles = append(titles, strings.TrimSuffix(name, path.Ext(name)))
		}
		err = templates.ExecuteTemplate(w, "index.html", titles)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {
	templates := template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html", "./tmpl/index.html"))
	validPath := regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

	http.HandleFunc("/", makeRootHandler(templates))
	http.HandleFunc("/view/", makeHandler(validPath, viewHandler(templates)))
	http.HandleFunc("/edit/", makeHandler(validPath, editHandler(templates)))
	http.HandleFunc("/save/", makeHandler(validPath, saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
