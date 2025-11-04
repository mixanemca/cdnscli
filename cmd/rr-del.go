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
	"github.com/spf13/cobra"
)

// rrDelCmd represents the add command
var rrDelCmd = &cobra.Command{
	Aliases: []string{"del", "rm", "remove", "unlink"},
	Args:    cobra.NoArgs,
	Use:     "delete",
	Short:   "Delete resource record from zone",
	Example: `  cdnscli rr delete --name www --zone example.com --type A --content 192.0.2.1`,
	Run:     rrDelCmdRun,
}

func init() {
	rrCmd.AddCommand(rrDelCmd)

	rrDelCmd.PersistentFlags().StringVarP(&content, "content", "c", "", "Comma separated IP address or domain name")
	if err := rrDelCmd.MarkPersistentFlagRequired("content"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "content", err)
	}
	rrDelCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "Zone name")
	if err := rrDelCmd.MarkPersistentFlagRequired("zone"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "zone", err)
	}
	rrDelCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "Resource record name")
	if err := rrDelCmd.MarkPersistentFlagRequired("name"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "name", err)
	}
	rrDelCmd.PersistentFlags().StringVarP(&rrtype, "type", "t", "", "Type of the resource record (A, CNAME)")
	if err := rrDelCmd.MarkPersistentFlagRequired("type"); err != nil {
		log.Fatalf("Failed to mark persistent flag %q as a required: %v", "type", err)
	}
}

func rrDelCmdRun(cmd *cobra.Command, args []string) {
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

	rrtype = strings.ToUpper(rrtype)

	ctx, cancel := context.WithTimeout(context.Background(), getTimeout())
	defer cancel()

	rr, err := a.Provider().GetRRByName(ctx, zone, name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = a.Provider().DeleteRR(ctx, zone, rr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	a.Printer().RecordDel(rr)
}
