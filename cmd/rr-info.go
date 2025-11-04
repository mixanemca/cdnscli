/*
Copyright © 2024-2025 Michael Bruskov <mixanemca@yandex.ru>

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

package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mixanemca/cdnscli/internal/app"
	"github.com/spf13/cobra"
)

// rrInfoCmd represents the info (details) command
var rrInfoCmd = &cobra.Command{
	Aliases: []string{"details"},
	Args:    cobra.NoArgs,
	Use:     "info",
	Short:   "Details for a single DNS record",
	Example: `  cdnscli rr info --name www --zone example.com`,
	Run:     rrInfoCmdRun,
}

func init() {
	rrCmd.AddCommand(rrInfoCmd)

	rrInfoCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "recource record name")
	if err := rrInfoCmd.MarkPersistentFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "name", err)
	}
	rrInfoCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "zone name")
	if err := rrInfoCmd.MarkPersistentFlagRequired("zone"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "zone", err)
	}

}

func rrInfoCmdRun(cmd *cobra.Command, args []string) {
	a, err := app.New(
		app.WithConfig(appConfig),
		app.WithOutputFormat(outputFormat),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
	defer cancel()

	rr, err := a.Provider().GetRRByName(ctx, zone, name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	a.Printer().RecordInfo(rr)
}
