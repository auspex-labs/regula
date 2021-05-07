// Copyright 2021 Fugue, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fugue/regula/pkg/loader"
	"github.com/spf13/cobra"
)

func NewShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [item]",
		Short: "Show debug information.",
		Long: `Show debug information.  Currently the available items are:
  regula-input [file..]   Show the JSON input being passed to regula`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Fprintf(os.Stderr, "Expected an item to show\n")
				os.Exit(1)
			}

			switch args[0] {
			case "regula-input":
				paths := args[1:]
				loadedFiles, err := loader.LoadPaths(loader.LoadPathsOptions{
					Paths: paths,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}

				bytes, err := json.MarshalIndent(loadedFiles.RegulaInput(), "", "  ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
				fmt.Println(string(bytes))

			default:
				fmt.Fprintf(os.Stderr, "Unknown item: %s\n", args[0])
				os.Exit(1)
			}
		},
	}

	return cmd
}

func init() {
	rootCmd.AddCommand(NewShowCommand())
}
