package zulu_test

import (
	"fmt"

	"github.com/zulucmd/zulu/v2"
)

func ExampleHookFuncE() {
	var rootCmd = &zulu.Command{
		Use:   "root [sub]",
		Short: "My root command",
		PersistentInitializeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PersistentInitializeE with args: %v\n", args)
			return nil
		},
		InitializeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd InitializeE with args: %v\n", args)
			return nil
		},
		PersistentPreRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PersistentPreRunE with args: %v\n", args)
			return nil
		},
		PreRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PreRunE with args: %v\n", args)
			return nil
		},
		RunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd RunE with args: %v\n", args)
			return nil
		},
		PostRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PostRunE with args: %v\n", args)
			return nil
		},
		PersistentPostRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PersistentPostRunE with args: %v\n", args)
			return nil
		},
		FinalizeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd FinalizeE with args: %v\n", args)
			return nil
		},
		PersistentFinalizeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside rootCmd PersistentFinalizeE with args: %v\n", args)
			return nil
		},
	}

	var subCmd = &zulu.Command{
		Use:   "sub [no options!]",
		Short: "My subcommand",
		PersistentInitializeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd PersistentInitializeE with args: %v\n", args)
			return nil
		},
		InitializeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd InitializeE with args: %v\n", args)
			return nil
		},
		PreRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd PreRunE with args: %v\n", args)
			return nil
		},
		RunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd RunE with args: %v\n", args)
			return nil
		},
		PostRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd PostRunE with args: %v\n", args)
			return nil
		},
		PersistentPostRunE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd PersistentPostRunE with args: %v\n", args)
			return nil
		},
		FinalizeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd FinalizeE with args: %v\n", args)
			return nil
		},
		PersistentFinalizeE: func(cmd *zulu.Command, args []string) error {
			fmt.Printf("Inside subCmd PersistentFinalizeE with args: %v\n", args)
			return nil
		},
	}

	rootCmd.AddCommand(subCmd)

	rootCmd.SetArgs([]string{""})
	_ = rootCmd.Execute()
	fmt.Println()
	rootCmd.SetArgs([]string{"sub", "arg1", "arg2"})
	_ = rootCmd.Execute()

	// Output:
	// Inside rootCmd PersistentInitializeE with args: []
	// Inside rootCmd InitializeE with args: []
	// Inside rootCmd PersistentPreRunE with args: []
	// Inside rootCmd PreRunE with args: []
	// Inside rootCmd RunE with args: []
	// Inside rootCmd PostRunE with args: []
	// Inside rootCmd PersistentPostRunE with args: []
	// Inside rootCmd FinalizeE with args: []
	// Inside rootCmd PersistentFinalizeE with args: []
	//
	// Inside subCmd PersistentInitializeE with args: []
	// Inside rootCmd PersistentInitializeE with args: []
	// Inside subCmd InitializeE with args: []
	// Inside rootCmd PersistentPreRunE with args: [arg1 arg2]
	// Inside subCmd PreRunE with args: [arg1 arg2]
	// Inside subCmd RunE with args: [arg1 arg2]
	// Inside subCmd PostRunE with args: [arg1 arg2]
	// Inside subCmd PersistentPostRunE with args: [arg1 arg2]
	// Inside rootCmd PersistentPostRunE with args: [arg1 arg2]
	// Inside subCmd FinalizeE with args: [arg1 arg2]
	// Inside subCmd PersistentFinalizeE with args: [arg1 arg2]
	// Inside rootCmd PersistentFinalizeE with args: [arg1 arg2]
}
