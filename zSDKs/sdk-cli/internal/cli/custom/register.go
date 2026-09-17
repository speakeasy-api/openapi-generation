// Package custom hosts hand-written commands that register into the generated CLI.
package custom

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Register receives the fully assembled root after every generated command has
// been attached. This file is generated once and is never overwritten on
// regeneration.
func Register(root *cobra.Command) {
	for _, hello := range root.Commands() {
		if hello.Name() != "hello" {
			continue
		}
		hello.Flags().String("greeting", "hello", "Greeting to print")
		hello.RunE = func(cmd *cobra.Command, args []string) error {
			greeting, err := cmd.Flags().GetString("greeting")
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s from custom code\n", greeting)
			return err
		}
	}
}
