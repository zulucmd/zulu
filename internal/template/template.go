package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"text/template"
)

func ParseFromFile(fs fs.FS, templateFile string, data interface{}, funcs template.FuncMap) (string, error) {
	f, err := fs.Open(templateFile)
	if err != nil {
		return "", fmt.Errorf("template: failed to open template file %q: %s", templateFile, err)
	}

	templateData, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("template: failed to read template file %q: %s", templateFile, err)
	}

	buf := new(bytes.Buffer)
	err = Parse(buf, string(templateData), data, funcs)
	if err != nil {
		return "", fmt.Errorf("template: failed to parse template: %s", err)
	}

	return buf.String(), nil
}

// Parse executes the given template text on data, writing the result to w.
func Parse(w io.Writer, text string, data interface{}, funcs template.FuncMap) error {
	t := template.New("top")
	t.Funcs(funcs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}
