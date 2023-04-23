---
weight: 210
---

## Man Pages

Generating man pages from a zulu command is incredibly easy. An example is as follows:

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
	header := &doc.GenManHeader{
		Title: "MINE",
		Section: "3",
	}
	err := doc.GenManTree(cmd, header, "/tmp")
	if err != nil {
		log.Fatal(err)
	}
}
```

That will get you a man page `/tmp/test.3`
