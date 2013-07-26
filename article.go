package main

import "strings"

type code_section interface {
	Code() bool
}

type para string

func (para) Code() bool {
	return false
}

type code string

func (code) Code() bool {
	return true
}

type Article struct {
	Title               string
	Intro               []para
	GoCode, ReflectCode []code_section
	Uses, Tags          []string
	slug                string
}

func (a *Article) Slug() string {
	if a.slug == "" {
		a.slug = strings.ToLower(strings.Replace(a.Title, " ", "-", -1))
	}
	return a.slug
}

type Articles []*Article

func (a *Articles) Add(ao *Article) {
	*a = append(*a, ao)
}

func (a Articles) Len() int {
	return len(a)
}

func (a Articles) Swap(i, j int) {
	a[j], a[i] = a[i], a[j]
}

func (a Articles) Less(i, j int) bool {
	return a[i].Title < a[j].Title
}
