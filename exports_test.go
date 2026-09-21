//nolint:testpackage // this is explicitly design to expose some functions in tests
package zulu

const (
	CompCmdName           = compCmdName
	CompCmdNoDescFlagName = compCmdDescFlagName
)

var StripFlags = stripFlags
var StringInSlice = stringInSlice

var ShellCompDirectiveMaxValue = shellCompDirectiveMaxValue

// InitCompleteCmd exposes initCompleteCmd so tests can mimic the __complete
// setup that ExecuteC performs.
func (c *Command) InitCompleteCmd(args []string) {
	c.initCompleteCmd(args)
}
