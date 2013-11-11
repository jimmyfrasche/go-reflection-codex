package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var fatal = log.Fatalln

func init() {
	log.SetFlags(0)
}

const dir = "/home/jimmy/code/go/src/github.com/jimmyfrasche/go-reflection-codex"

func mksorter(of map[string]*Articles) func(string) *Articles {
	return func(key string) *Articles {
		return of[key]
	}
}

func dashify(s string) string {
	s = strings.ToLower(s)
	return strings.Replace(strings.Replace(s, " ", "-", -1), ".", "-", -1)
}

func sort_and_keys(articles map[string]*Articles) (out []string) {
	for k := range articles {
		sort.Sort(articles[k])
		out = append(out, k)
	}
	sort.Strings(out)
	return
}

func main() {
	entries := filepath.Join(dir, "entries")
	d, err := os.Open(entries)
	if err != nil {
		fatal(err)
	}

	if err = d.Chdir(); err != nil {
		fatal(err)
	}
	names, err := d.Readdirnames(-1)
	if err != nil {
		fatal(err)
	}

	var articles Articles

	title2file := map[string]string{}
	for _, name := range names {
		f, err := ioutil.ReadFile(name)
		if err != nil {
			fatal(name+":", err)
		}

		article, err := Parse(f)
		if err != nil {
			fatal(name+":", err)
		}

		//make sure article names are unique
		if other := title2file[article.Title]; other != "" {
			fatal(article.Title, "defined in", name, "and", other)
		}
		title2file[article.Title] = name

		articles.Add(article)
	}

	bytags := map[string]*Articles{}
	byuses := map[string]*Articles{}

	for _, article := range articles {
		for _, tag := range article.Tags {
			if bytags[tag] == nil {
				bytags[tag] = &Articles{}
			}
			bytags[tag].Add(article)
		}
		for _, use := range article.Uses {
			if byuses[use] == nil {
				byuses[use] = &Articles{}
			}
			byuses[use].Add(article)
		}
	}

	sort.Sort(articles)

	tagsort := sort_and_keys(bytags)
	usesort := sort_and_keys(byuses)

	data := map[string]interface{}{
		"Articles":  articles,
		"TopicSort": tagsort,
		"UsesSort":  usesort,
	}

	funcs := template.FuncMap{
		"topicsOf": mksorter(bytags),
		"usesOf":   mksorter(byuses),
		"dashify":  dashify,
	}
	const tmplname = "reflection-codex.html.template"
	tmpl := filepath.Join(dir, tmplname)
	html, err := template.New("tmpl").Funcs(funcs).ParseFiles(tmpl)
	if err != nil {
		fatal(err)
	}
	if err = html.ExecuteTemplate(os.Stdout, tmplname, data); err != nil {
		fatal(err)
	}
}
