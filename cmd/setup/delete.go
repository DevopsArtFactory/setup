/*
Copyright © 2020 DevopsArtFactory gwonsoo.lee@gmail.com

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
	"os"

	"github.com/GwonsooLee/kubenx/pkg/color"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete key",
	Long: "Delete key",
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if err := DeleteCmd(); err != nil {
			color.Red.Fprintln(os.Stdout, err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
