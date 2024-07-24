/*
Copyright © 2024 Michael Bruskov <mixanemca@yandex.ru>

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
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cfdnscli/app"
	"github.com/spf13/cobra"
)

// rrListCmd represents the list (ls) command
var rrListCmd = &cobra.Command{
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Use:     "list",
	Short:   "List of zone resource records",
	Example: `  cfdnscli rr list --zone example.com`,
	Run:     rrListCmdRun,
}

func init() {
	rrCmd.AddCommand(rrListCmd)

	rrListCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "zone name")
	rrListCmd.MarkPersistentFlagRequired("zone")
}

func rrListCmdRun(cmd *cobra.Command, args []string) {
	a, err := app.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	recs, err := a.Zones().ListRecords(ctx, zone, cloudflare.ListDNSRecordsParams{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(recs)
}
