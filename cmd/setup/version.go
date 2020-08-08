/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package setup

import (
	"github.com/GwonsooLee/kubenx/pkg/color"
	"os"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Check version of setup",
	Long:    "Check version of setup",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		version := "1.1.0"
		color.Blue.Fprintf(os.Stdout, "v%s", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
