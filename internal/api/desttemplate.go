package api

import (
	"bytes"
	"strings"
	"text/template"
)

type destPathParams struct {
	DatasetFolder string
	SourceFolder  string
	Pid           string
	PidShort      string
	PidPrefix     string
	PidEncoded    string
	Username      string
}

type DestinationTemplate struct {
	template *template.Template
}

func NewDestinationTemplate(templateString string) (DestinationTemplate, error) {
	var dt DestinationTemplate
	dt.template = template.New("dataset destination path template").Funcs(
		template.FuncMap{
			"replace": func(s string, query string, repl string) string {
				return strings.ReplaceAll(s, query, repl)
			},
		},
	)

	var err error
	dt.template, err = dt.template.Parse(templateString)
	return dt, err
}

func (dt *DestinationTemplate) Execute(data destPathParams) (string, error) {
	buffer := bytes.Buffer{}
	err := dt.template.Execute(&buffer, data)
	return buffer.String(), err
}
