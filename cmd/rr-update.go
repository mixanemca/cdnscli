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
	"log"
	"os"

	"github.com/mixanemca/cdnscli/internal/app"
	"github.com/spf13/cobra"
)

// rrUpdateCmd represents the update command
var rrUpdateCmd = &cobra.Command{
	Aliases: []string{"change", "move", "mv", "patch"},
	Args:    cobra.NoArgs,
	Use:     "update",
	Short:   "Update an existing DNS record",
	Example: `  cdnscli rr update --name www --zone example.com --type A --content 192.0.2.1`,
	Run:     rrUpdateCmdRun,
}

func init() {
	rrCmd.AddCommand(rrUpdateCmd)

	rrUpdateCmd.PersistentFlags().StringVarP(&content, "content", "c", "", "Comma separated IP address or domain name")
	if err := rrUpdateCmd.MarkPersistentFlagRequired("content"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "content", err)
	}
	rrUpdateCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "recource record name")
	if err := rrUpdateCmd.MarkPersistentFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "name", err)
	}
	// rrUpdateCmd.PersistentFlags().BoolVarP(&proxied, "proxied", "p", false, "Whether the record is receiving the performance and security benefits of Cloudflare")
	rrUpdateCmd.PersistentFlags().StringVarP(&rrtype, "type", "t", "", "Type of the resource record (A, CNAME)")
	if err := rrUpdateCmd.MarkPersistentFlagRequired("type"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "type", err)
	}
	// rrUpdateCmd.PersistentFlags().IntVarP(&ttl, "ttl", "l", 1800, "The time to live of the resource record in seconds")
	rrUpdateCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "zone name")
	if err := rrUpdateCmd.MarkPersistentFlagRequired("zone"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "zone", err)
	}
}

func rrUpdateCmdRun(cmd *cobra.Command, args []string) {
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
	rr.Content = content
	rr.Type = rrtype
	// rr.TTL = ttl
	// rr.Proxied = cloudflare.BoolPtr(proxied)

	updated, err := a.Provider().UpdateRR(ctx, zone, rr)
	if err != nil {
		return
	}

	a.Printer().RecordUpdate(updated)
}
