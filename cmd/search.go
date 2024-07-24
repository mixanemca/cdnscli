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
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/mixanemca/cfdnscli/app"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search resource records",
	Example: "  cfdnscli search --zone example.com --content 192.0.2.1",
	Run:     searchCmdRun,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.PersistentFlags().StringVarP(&content, "content", "c", "", "the content string to search for")
	searchCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "the resourse record name to search for")
	// searchCmd.PersistentFlags().IntVarP(&max, "max", "m", 10, "maximum number of entries to return")
	// searchCmd.PersistentFlags().StringVarP(&rrtype, "type", "t", "A", "type of resorce record to search for")
	searchCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "the zone name")
	searchCmd.MarkPersistentFlagRequired("zone")
}

func searchCmdRun(cmd *cobra.Command, args []string) {
	if len(content) == 0 && len(name) == 0 {
		fmt.Println("ERROR: you must specify one of the search parameters - content or name")
		os.Exit(1)
	}

	a, err := app.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	if len(name) > 0 {
		name = strings.Join([]string{name, zone}, ".")
	}

	results, err := a.Zones().ListRecords(ctx, zone, cloudflare.ListDNSRecordsParams{
		Content: content,
		Name:    name,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(results)
}
