package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bitfield/script"
)

func main() {
	ctx := kong.Parse(&cmd{},
		kong.Description("Tool for converting config file to concise version."),
		kong.ShortUsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}

type cmd struct {
	Config string `arg:"" help:"config version"`
	Schema string `arg:"" help:"schema version"`
}

func (c cmd) source() string { return c.Config + "/reference/.golangci.yml" }
func (c cmd) output() string { return c.Config + "/.golangci.yml" }
func (c cmd) schema() string { return "jsonschema/golangci.v" + c.Schema + ".jsonschema.json" }

func (c cmd) Validate() error {
	reVersion := regexp.MustCompile(`^[0-9.]+$`)
	errNoSource := script.IfExists(c.source()).Error()
	errNoSchema := script.IfExists(c.schema()).Error()
	switch {
	case !reVersion.MatchString(c.Config):
		return fmt.Errorf("<config>: invalid version %q", c.Config)
	case !reVersion.MatchString(c.Schema):
		return fmt.Errorf("<schema>: invalid version %q", c.Schema)
	case errNoSource != nil:
		return fmt.Errorf("<config>: wrong version %q: %w", c.Config, errNoSource)
	case errNoSchema != nil:
		return fmt.Errorf("<schema>: wrong version %q: %w", c.Schema, errNoSchema)
	}
	return nil
}

func (c cmd) Run() error {
	re := regexp.MustCompile
	header := "# Origin: https://github.com/powerman/golangci-lint-strict version " + c.Config + "\n"
	cmdLineValidate := fmt.Sprintf("jv %s %s", c.schema(), c.output())

	_, err := newPipe(nil).
		Exec(`yq --prettyPrint --output-format yaml ` + c.source()).
		RejectRegexp(re(`^\s*(?:#|$)`)).
		WriteFile(c.output()) // jv can't read from stdin, so we need a file.
	rules, _ := newPipe(err).
		Exec(cmdLineValidate).
		Match(`: got null, want`).
		ReplaceRegexp(re(`.*'(.*)'.*`), `del($1)`).
		Replace(`/`, `.`).
		Slice()
	rule := strings.Join(append(rules, "."), "|")
	lines, err := newPipe(err).
		Exec(fmt.Sprintf(`yq "%s" %s`, rule, c.output())).
		RejectRegexp(re(`^(?:.*: {}|\s*)$`)).
		String()
	_, err = newPipe(err).Echo(header + lines).WriteFile(c.output())
	_, _ = newPipe(err).Exec(cmdLineValidate).Stdout()

	if err != nil {
		_ = os.Remove(c.output())
	}
	return err
}

func newPipe(err error) *script.Pipe {
	return script.NewPipe().WithStderr(os.Stderr).WithError(err)
}
