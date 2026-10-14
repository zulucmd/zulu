package zulu_test

import (
	"errors"
	"fmt"

	"github.com/zulucmd/zulu/v2"
)

// A validator can reject the invocation without caring where the rejection came
// from, and the exit code follows from the classification. Both streams are
// silenced so the example prints only what it chose to.
func ExampleUsage() {
	root := &zulu.Command{
		Use:           "deploy",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(_ *zulu.Command, args []string) error {
			if len(args) != 1 {
				return zulu.Usagef("deploy takes one environment, not %d", len(args))
			}

			return nil
		},
		RunE: func(_ *zulu.Command, _ []string) error { return nil },
	}
	root.SetArgs([]string{"staging", "prod"})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		fmt.Println("exit code:", zulu.ExitCode(err))
	}

	// A type of the caller's own can be classified as well, and stays reachable
	// for an errors.As check.
	err := zulu.Usage(&fieldError{field: "name"})

	var field *fieldError
	fmt.Println("usage:", zulu.IsUsage(err), "type kept:", errors.As(err, &field))

	// Output:
	// deploy takes one environment, not 2
	// exit code: 2
	// usage: true type kept: true
}
