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

// rrAddCmd represents the add command
var rrAddCmd = &cobra.Command{
	Aliases: []string{"new", "create"},
	Args:    cobra.NoArgs,
	Use:     "add",
	Short:   "Add resource record to zone",
	Example: `  cfdnscli rr add --name www --zone example.com --type A --ttl 400 --content 10.0.0.1`,
	Run:     rrAddCmdRun,
}

func init() {
	rrCmd.AddCommand(rrAddCmd)

	rrAddCmd.PersistentFlags().StringVarP(&content, "content", "c", "", "Comma separated IP address or domain name")
	rrAddCmd.MarkPersistentFlagRequired("content")
	rrAddCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "Zone name")
	rrAddCmd.MarkPersistentFlagRequired("zone")
	rrAddCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Resource record name")
	rrAddCmd.MarkPersistentFlagRequired("name")
	rrAddCmd.PersistentFlags().IntVarP(&ttl, "ttl", "l", 1800, "The time to live of the resource record in seconds")
	rrAddCmd.PersistentFlags().StringVarP(&rrtype, "type", "t", "", "Type of the resource record (A, CNAME)")
	rrAddCmd.MarkPersistentFlagRequired("type")
}

func rrAddCmdRun(cmd *cobra.Command, args []string) {
	a, err := app.New()
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

	params := cloudflare.CreateDNSRecordParams{
		Content: content,
		Name:    name,
		TTL:     ttl,
		Type:    rrtype,
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	rr, err := a.Zones().AddRR(ctx, zone, params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(rr)
}
