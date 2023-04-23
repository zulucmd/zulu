package doc_test

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zulucmd/zflag/v2"

	"github.com/zulucmd/zulu"
	"github.com/zulucmd/zulu/doc"
)

func translate(in string) string {
	return strings.ReplaceAll(in, "-", "\\-")
}

func TestGenManDoc(t *testing.T) {
	header := &doc.GenManHeader{
		Title:   "Project",
		Section: "2",
	}

	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenMan(echoCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	// Make sure parent has - in CommandPath() in SEE ALSO:
	parentPath := echoCmd.Parent().CommandPath()
	dashParentPath := strings.ReplaceAll(parentPath, " ", "-")
	expected := translate(dashParentPath)
	expected = expected + "(" + header.Section + ")"
	assertContains(t, output, expected)

	assertContains(t, output, translate(echoCmd.Name()))
	assertContains(t, output, translate(echoCmd.Name()))
	assertContains(t, output, "boolone")
	assertContains(t, output, "rootflag")
	assertContains(t, output, translate(rootCmd.Name()))
	assertContains(t, output, translate(echoSubCmd.Name()))
	assertNotContains(t, output, translate(deprecatedCmd.Name()))
	assertContains(t, output, translate("Auto generated"))
}

func TestGenManNoHiddenParents(t *testing.T) {
	header := &doc.GenManHeader{
		Title:   "Project",
		Section: "2",
	}

	// We generate on a subcommand so we have both subcommands and parents
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := doc.GenMan(echoCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	// Make sure parent has - in CommandPath() in SEE ALSO:
	parentPath := echoCmd.Parent().CommandPath()
	dashParentPath := strings.ReplaceAll(parentPath, " ", "-")
	expected := translate(dashParentPath)
	expected = expected + "(" + header.Section + ")"
	assertContains(t, output, expected)

	assertContains(t, output, translate(echoCmd.Name()))
	assertContains(t, output, translate(echoCmd.Name()))
	assertContains(t, output, "boolone")
	assertNotContains(t, output, "rootflag")
	assertContains(t, output, translate(rootCmd.Name()))
	assertContains(t, output, translate(echoSubCmd.Name()))
	assertNotContains(t, output, translate(deprecatedCmd.Name()))
	assertContains(t, output, translate("Auto generated"))
	assertNotContains(t, output, "OPTIONS INHERITED FROM PARENT COMMANDS")
}

func TestGenManNoGenTag(t *testing.T) {
	echoCmd.DisableAutoGenTag = true
	defer func() { echoCmd.DisableAutoGenTag = false }()

	header := &doc.GenManHeader{
		Title:   "Project",
		Section: "2",
	}

	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenMan(echoCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	unexpected := translate("#HISTORY")
	assertNotContains(t, output, unexpected)
	unexpected = translate("Auto generated by zulucmd/zulu")
	assertNotContains(t, output, unexpected)
}

func TestGenManNoGenTagWithDisabledParent(t *testing.T) {
	// We set the flag on a parent to check it is used in its descendance
	rootCmd.DisableAutoGenTag = true
	defer func() {
		echoCmd.DisableAutoGenTag = false
		rootCmd.DisableAutoGenTag = false
	}()

	header := &doc.GenManHeader{
		Title:   "Project",
		Section: "2",
	}

	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenMan(echoCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	unexpected := translate("#HISTORY")
	assertNotContains(t, output, unexpected)
	unexpected = translate("Auto generated by zulucmd/zulu")
	assertNotContains(t, output, unexpected)
}

func TestGenManSeeAlso(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	aCmd := &zulu.Command{Use: "aaa", RunE: emptyRun, Hidden: true} // #229
	bCmd := &zulu.Command{Use: "bbb", RunE: emptyRun}
	cCmd := &zulu.Command{Use: "ccc", RunE: emptyRun}
	rootCmd.AddCommand(aCmd, bCmd, cCmd)

	buf := new(bytes.Buffer)
	header := &doc.GenManHeader{}
	if err := doc.GenMan(rootCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(buf)

	if err := assertLineFound(scanner, ".SH SEE ALSO"); err != nil {
		t.Fatalf("Couldn't find SEE ALSO section header: %v", err)
	}
	if err := assertNextLineEquals(scanner, ".PP"); err != nil {
		t.Fatalf("First line after SEE ALSO wasn't break-indent: %v", err)
	}
	if err := assertNextLineEquals(scanner, `\fBroot-bbb(1)\fP, \fBroot-ccc(1)\fP, \fBroot-completion(1)\fP`); err != nil {
		t.Fatalf("Second line after SEE ALSO wasn't correct: %v", err)
	}
}

func TestManPrintFlagsHidesShortDeprecated(t *testing.T) {
	c := &zulu.Command{}
	c.Flags().String("foo", "default", "Foo flag", zflag.OptShorthand('f'), zflag.OptShorthandDeprecated("don't use it no more"))

	buf := new(bytes.Buffer)
	doc.ManPrintFlags(buf, c.Flags())

	got := buf.String()
	expected := "**--foo**=\"default\"\n\tFoo flag\n\n"
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestGenManCommands(t *testing.T) {
	header := &doc.GenManHeader{
		Title:   "Project",
		Section: "2",
	}

	// Root command
	buf := new(bytes.Buffer)
	if err := doc.GenMan(rootCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, ".SH COMMANDS")
	assertMatch(t, output, "\\\\fBecho\\\\fP\n[ \t]+Echo anything to the screen\n[ \t]+See \\\\fBroot\\-echo\\(2\\)\\\\fP\\\\&\\.")
	assertNotContains(t, output, ".PP\n\\fBprint\\fP\n")

	// Echo command
	buf = new(bytes.Buffer)
	if err := doc.GenMan(echoCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output = buf.String()

	assertContains(t, output, ".SH COMMANDS")
	assertMatch(t, output, "\\\\fBtimes\\\\fP\n[ \t]+Echo anything to the screen more times\n[ \t]+See \\\\fBroot\\-echo\\-times\\(2\\)\\\\fP\\\\&\\.")
	assertMatch(t, output, "\\\\fBechosub\\\\fP\n[ \t]+second sub command for echo\n[ \t]+See \\\\fBroot\\-echo\\-echosub\\(2\\)\\\\fP\\\\&\\.")
	assertNotContains(t, output, ".PP\n\\fBdeprecated\\fP\n")

	// Time command as echo's subcommand
	buf = new(bytes.Buffer)
	if err := doc.GenMan(timesCmd, header, buf); err != nil {
		t.Fatal(err)
	}
	output = buf.String()

	assertNotContains(t, output, ".SH COMMANDS")
}

func TestGenManTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}
	header := &doc.GenManHeader{Section: "2"}
	tmpdir, err := os.MkdirTemp("", "test-gen-man-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %s", err.Error())
	}
	defer os.RemoveAll(tmpdir)

	if err := doc.GenManTree(c, header, tmpdir); err != nil {
		t.Fatalf("GenManTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.2")); err != nil {
		t.Fatalf("Expected file 'do.2' to exist")
	}

	if header.Title != "" {
		t.Fatalf("Expected header.Title to be unmodified")
	}
}

func assertLineFound(scanner *bufio.Scanner, expectedLine string) error {
	for scanner.Scan() {
		line := scanner.Text()
		if line == expectedLine {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan failed: %s", err)
	}

	return fmt.Errorf("hit EOF before finding %v", expectedLine)
}

func assertNextLineEquals(scanner *bufio.Scanner, expectedLine string) error {
	if scanner.Scan() {
		line := scanner.Text()
		if line == expectedLine {
			return nil
		}
		return fmt.Errorf("got %v, not %v", line, expectedLine)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan failed: %v", err)
	}

	return fmt.Errorf("hit EOF before finding %v", expectedLine)
}

func BenchmarkGenManToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := doc.GenMan(rootCmd, nil, file); err != nil {
			b.Fatal(err)
		}
	}
}
