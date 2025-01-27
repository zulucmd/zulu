---
weight: 250
---

## AsciiDoc Docs

Generating AsciiDoc pages from a zulu command is incredibly easy. An example is as follows:

```go
package main

import (
	"log"

	"{{< param go_import_package >}}"
	"{{< param go_import_package >}}/doc"
)

func main() {
	cmd := &zulu.Command{
		Use:   "test",
		Short: "my test program",
	}
	err := doc.GenAsciidocTree(cmd, "/tmp")
	if err != nil {
		log.Fatal(err)
	}
}
```

That will get you a AsciiDoc document `/tmp/test.adoc`

### Generate for the entire command tree

This program can actually generate docs for the kubectl command in the kubernetes project

```go
package main

import (
	"log"
	"io"
	"os"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"{{< param go_import_package >}}/doc"
)

func main() {
	kubectl := cmd.NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, io.Discard, io.Discard)
	err := doc.GenAsciidocTree(kubectl, "./")
	if err != nil {
		log.Fatal(err)
	}
}
```

This will generate a whole series of files, one for each command in the tree, in the directory specified (in this case "./")

### Generate for a single command

You may wish to have more control over the output, or only generate for a single command, instead of the entire command tree. If this is the case you may prefer to `GenAsciidoc` instead of `GenAsciidocTree`

```go
	out := new(bytes.Buffer)
	err := doc.GenAsciidoc(cmd, out)
	if err != nil {
		log.Fatal(err)
	}
```

This will write the AsciiDoc doc for ONLY "cmd" into the out, buffer.

### Customize the output

Both `GenAsciidoc` and `GenAsciidocTree` have alternate versions with callbacks to get some control of the output:

```go
func GenAsciidocTreeCustom(cmd *Command, dir string, filePrepender, linkHandler func(string) string) error {
	//...
}
```

```go
func GenAsciidocCustom(cmd *Command, out *bytes.Buffer, linkHandler func(string) string) error {
	//...
}
```

The `filePrepender` will prepend the return value given the full filepath to the rendered Asciidoc file. A common use case is to add front matter to use the generated documentation with [Hugo](https://gohugo.io/):

```go
const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---
`

filePrepender := func(filename string) string {
	now := time.Now().Format(time.RFC3339)
	name := filepath.Base(filename)
	base := strings.TrimSuffix(name, path.Ext(name))
	url := "/commands/" + strings.ToLower(base) + "/"
	return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
}
```

The `linkHandler` can be used to customize the rendered internal links to the commands, given a filename:

```go
linkHandler := func(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	return "/commands/" + strings.ToLower(base) + "/"
}
```
