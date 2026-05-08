/*
Copyright © 2024-2026 Michael Bruskov <mixanemca@yandex.ru>

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
	"sort"

	"github.com/mixanemca/cdnscli/internal/app"
	"github.com/mixanemca/cdnscli/internal/models"
	"github.com/spf13/cobra"
)

// zoneListCmd represents the list command
var zoneListCmd = &cobra.Command{
	Aliases: []string{"ls"},
	Use:     "list",
	Short:   "Lists zones on an account. Optionally takes a name of zone to filter against.",
	Example: "  cdnscli zone list",
	Run:     zoneListRun,
}

func init() {
	zoneCmd.AddCommand(zoneListCmd)

	zoneListCmd.PersistentFlags().StringVarP(&name, "name", "n", "", "name of zone to filter against")
}

func zoneListRun(cmd *cobra.Command, args []string) {
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

	var allZones []models.Zone
	for _, pName := range a.ProviderNames() {
		p, err := a.GetProvider(pName)
		if err != nil {
			continue
		}
		var zones []models.Zone
		if len(name) > 0 {
			zones, err = p.ListZonesByName(ctx, name)
		} else {
			zones, err = p.ListZones(ctx)
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		displayName := a.ProviderDisplayName(pName)
		for i := range zones {
			zones[i].ProviderName = displayName
		}
		allZones = append(allZones, zones...)
	}

	sort.Slice(allZones, func(i, j int) bool { return allZones[i].Name < allZones[j].Name })
	a.Printer().ZonesList(allZones)
}
