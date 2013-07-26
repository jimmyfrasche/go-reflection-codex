package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type Parser struct {
	buf *bytes.Buffer
	err error
}

func (p *Parser) line() (line string, trimmed string, skip bool) {
	if p.err != nil {
		return "", "", true
	}
	line, err := p.buf.ReadString('\n')
	if err != nil {
		p.err = err
		return "", "", true
	}
	sline := strings.TrimSpace(line)
	if sline == "" {
		return sline, sline, false
	}
	if strings.HasPrefix(line, "==") {
		return "", "", true
	}
	return line, sline, false
}

//skip all preceeding empty lines
func (p *Parser) nextLine() (string, string, bool) {
	for {
		line, sline, skip := p.line()
		if skip {
			return line, sline, skip
		}
		if sline != "" {
			return line, sline, skip
		}
	}
}

func (p *Parser) Title() (string, error) {
	_, sline, skip := p.nextLine()
	if skip {
		return "", errors.New("No title")
	}
	return sline, nil
}

var _no_tags = errors.New("No tags")

func (p *Parser) Tags() (tags []string, err error) {
	_, sline, skip := p.nextLine()
	if skip {
		return nil, _no_tags
	}
	for _, s := range strings.Split(sline, ",") {
		if s = strings.TrimSpace(s); s != "" {
			tags = append(tags, s)
		}
	}
	if len(tags) == 0 {
		return nil, _no_tags
	}
	return tags, nil
}

func notempty(s []string) bool {
	for _, si := range s {
		if len(si) > 0 {
			return true
		}
	}
	return false
}

func (p *Parser) Intro() (paras []para) {
	_, line, skip := p.nextLine()
	if skip {
		return
	}
	acc := []string{line}
	finalize := func() {
		if notempty(acc) {
			paras = append(paras, para(strings.Join(acc, " ")))
			acc = []string{}
		}
	}
	for {
		_, line, skip = p.line()
		if skip {
			break
		}
		if line == "" {
			finalize()
		} else {
			acc = append(acc, line)
		}
	}
	finalize()
	return
}

func (p *Parser) Code() (cs []code_section, err error) {
	line, _, skip := p.nextLine()
	if skip {
		return nil, errors.New("Empty code section")
	}
	iscode := func() bool {
		return len(line) > 0 && line[0] == '\t'
	}
	var pacc, cacc []string
	close_para := func() {
		if notempty(pacc) {
			cs = append(cs, para(strings.Join(pacc, " ")))
			pacc = []string{}
		}
	}
	close_code := func() {
		if notempty(cacc) {
			new := strings.Join(cacc, "")
			new = strings.TrimRight(new, "\n")
			cs = append(cs, code(new))
			cacc = []string{}
		}
	}
	wascode := false
	lastwascode := iscode()

	if lastwascode {
		cacc = append(cacc, line[1:])
		wascode = true
	} else {
		pacc = append(pacc, line)
	}
	for {
		line, _, skip = p.line()
		if skip {
			break
		}
		if iscode() {
			if !lastwascode {
				close_para()
			}
			lastwascode = true
			wascode = true
			cacc = append(cacc, line[1:])
		} else {
			if lastwascode {
				close_code()
			}
			lastwascode = false
			pacc = append(pacc, line)
			if len(line) == 0 {
				close_para()
			}
		}
	}
	if lastwascode {
		close_code()
	} else {
		close_para()
	}
	if !wascode {
		return nil, errors.New("No code in code section")
	}
	return
}

func Parse(file []byte) (article *Article, err error) {
	article = &Article{}

	p := &Parser{
		buf: bytes.NewBuffer(file),
	}

	title, err := p.Title()
	if err != nil {
		return nil, err
	}
	article.Title = title

	tags, err := p.Tags()
	if err != nil {
		return nil, err
	}
	article.Tags = tags

	article.Intro = p.Intro()

	go_code, err := p.Code()
	if err != nil {
		return nil, err
	}
	if len(go_code) == 0 {
		return nil, errors.New("No Go code segment")
	}
	article.GoCode = go_code

	reflect_code, err := p.Code()
	if err != nil {
		return nil, err
	}
	if len(reflect_code) == 0 {
		return nil, errors.New("No Go code segment")
	}
	article.ReflectCode = reflect_code

	if p.err != nil && p.err != io.EOF {
		return nil, p.err
	}

	//everything parsed fine so let's see if there are any specifically mentioned
	//reflect funcs/types
	uses := []string{}
	for str, bstr := range topics {
		if bytes.Contains(file, bstr) {
			uses = append(uses, str)
		}
	}
	article.Uses = uses

	return
}
