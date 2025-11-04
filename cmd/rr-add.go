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
	"strings"

	"github.com/mixanemca/cdnscli/internal/app"
	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/spf13/cobra"
)

// rrAddCmd represents the add command
var rrAddCmd = &cobra.Command{
	Aliases: []string{"new", "create"},
	Args:    cobra.NoArgs,
	Use:     "add",
	Short:   "Add resource record to zone",
	Example: `  cdnscli rr add --name www --zone example.com --type A --ttl 400 --content 192.0.2.1`,
	Run:     rrAddCmdRun,
}

func init() {
	rrCmd.AddCommand(rrAddCmd)

	rrAddCmd.PersistentFlags().StringVarP(&content, "content", "c", "", "Comma separated IP address or domain name")
	if err := rrAddCmd.MarkPersistentFlagRequired("content"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "content", err)
	}
	rrAddCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "Zone name")
	if err := rrAddCmd.MarkPersistentFlagRequired("zone"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "zone", err)
	}
	rrAddCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Resource record name")
	if err := rrAddCmd.MarkPersistentFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "name", err)
	}
	rrAddCmd.PersistentFlags().BoolVarP(&proxied, "proxied", "p", false, "Whether the record is receiving the performance and security benefits of Cloudflare")
	rrAddCmd.PersistentFlags().IntVarP(&ttl, "ttl", "l", 1800, "The time to live of the resource record in seconds")
	rrAddCmd.PersistentFlags().StringVarP(&rrtype, "type", "t", "", "Type of the resource record (A, CNAME)")
	if err := rrAddCmd.MarkPersistentFlagRequired("type"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "type", err)
	}
}

func rrAddCmdRun(cmd *cobra.Command, args []string) {
	a, err := app.New(
		app.WithConfig(appConfig),
		app.WithOutputFormat(outputFormat),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check that name not FQDN
	if strings.Contains(name, zone) {
		fmt.Printf("ERROR: Name (%s) must not be a FQDN. Without domain %s\n", name, zone)
		os.Exit(1)
	}
	// name = hostname + example.com
	name = strings.Join([]string{name, zone}, ".")

	rrtype = strings.ToUpper(rrtype)

	params := models.CreateDNSRecordParams{
		Content:  content,
		Name:     name,
		Proxied:  proxied,
		TTL:      ttl,
		Type:     rrtype,
		ZoneName: zone,
	}

	ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
	defer cancel()

	rr, err := a.Provider().AddRR(ctx, zone, params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	a.Printer().RecordAdd(rr)
}
