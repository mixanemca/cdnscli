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

package prettyprint

import (
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// TextPrinter prints in human-readable format.
type TextPrinter struct{}

// ZonesList prints list of DNS zones.
func (pp *TextPrinter) ZonesList(zones []cloudflare.Zone) {
	var fields strings.Builder
	for _, z := range zones {
		fields.WriteString(fmt.Sprintf("ID: %s\n", z.ID))
		fields.WriteString(fmt.Sprintf("Name: %s\n", z.Name))
		fields.WriteString(fmt.Sprintf("Name Servers: %s\n", strings.Join(z.NameServers, ", ")))
		fields.WriteString(fmt.Sprintf("Status: %s\n", z.Status))
	}
	fmt.Print(fields.String())
}
