package template

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"text/template"
)

func ParseFromFile(fsys fs.FS, templateFile string, data any, funcs template.FuncMap) (string, error) {
	t, err := template.New("root").Funcs(funcs).ParseFS(fsys, templateFile)
	if err != nil {
		return "", fmt.Errorf("template: failed to parse template file %q: %w", templateFile, err)
	}

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, filepath.Base(templateFile), data)
	if err != nil {
		return "", fmt.Errorf("template: failed to parse template: %w", err)
	}

	return buf.String(), nil
}

// Parse executes the given template text on data, writing the result to w.
func Parse(w io.Writer, text string, data any, funcs template.FuncMap) error {
	t := template.New("top")
	t.Funcs(funcs)
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}
