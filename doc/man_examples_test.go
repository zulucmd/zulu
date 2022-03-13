package doc_test

import (
	"bytes"
	"fmt"

	"github.com/gowarden/zulu"
	"github.com/gowarden/zulu/doc"
	"github.com/gowarden/zulu/internal/util"
)

func ExampleGenManTree() {
	cmd := &zulu.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	util.CheckErr(doc.GenManTree(cmd, header, "/tmp"))
}

func ExampleGenMan() {
	cmd := &zulu.Command{
		Use:   "test",
		Short: "my test program",
	}
	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}
	out := new(bytes.Buffer)
	util.CheckErr(doc.GenMan(cmd, header, out))
	fmt.Print(out.String())
}
