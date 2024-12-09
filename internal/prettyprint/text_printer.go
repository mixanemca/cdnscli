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

	"github.com/mixanemca/cfdnscli/internal/models"
)

// TextPrinter prints in human-readable format.
type TextPrinter struct{}

// ZonesList prints list of DNS zones.
func (pp *TextPrinter) ZonesList(zones []models.Zone) {
	var fields strings.Builder
	for _, z := range zones {
		fields.WriteString(fmt.Sprintf("ID: %s\n", z.ID))
		fields.WriteString(fmt.Sprintf("Name: %s\n", z.Name))
		fields.WriteString(fmt.Sprintf("Name Servers: %s\n", strings.Join(z.NameServers, ", ")))
		fields.WriteString(fmt.Sprintf("Status: %s\n", z.Status))
	}
	fmt.Print(fields.String())
}

// RecordsList prints list of DNS resource records.
func (pp *TextPrinter) RecordsList(rrset []models.DNSRecord) {
	var fields strings.Builder
	for _, rr := range rrset {
		fields.WriteString(fmt.Sprintf("ID: %s\n", rr.ID))
		fields.WriteString(fmt.Sprintf("Name: %s\n", rr.Name))
		fields.WriteString(fmt.Sprintf("TTL: %d\n", rr.TTL))
		fields.WriteString(fmt.Sprintf("Type: %s\n", rr.Type))
		fields.WriteString(fmt.Sprintf("Proxied: %t\n", rr.Proxied))
		fields.WriteString(fmt.Sprintf("Content: %s\n", rr.Content))
	}
	fmt.Print(fields.String())
}

// RecordInfo displays information about a specified DNS resource record.
func (pp *TextPrinter) RecordInfo(rr models.DNSRecord) {
	var fields strings.Builder

	fields.WriteString(fmt.Sprintf("ID: %s\n", rr.ID))
	fields.WriteString(fmt.Sprintf("Name: %s\n", rr.Name))
	fields.WriteString(fmt.Sprintf("TTL: %d\n", rr.TTL))
	fields.WriteString(fmt.Sprintf("Type: %s\n", rr.Type))
	fields.WriteString(fmt.Sprintf("Proxied: %t\n", rr.Proxied))
	fields.WriteString(fmt.Sprintf("Content: %s\n", rr.Content))

	fmt.Print(fields.String())

}

// RecordAdd displays information about a new DNS resource record.
func (pp *TextPrinter) RecordAdd(rr models.DNSRecord) {
	var fields strings.Builder

	fields.WriteString(fmt.Sprintf("New resource record %q was been added with ID %q\n",
		rr.Name,
		rr.ID,
	))

	fmt.Print(fields.String())
}

// RecordDel displays information about a deleted DNS recource record.
func (pp *TextPrinter) RecordDel(rr models.DNSRecord) {
	fmt.Printf("DNS resource record %s successfully deleted\n", rr.Name)
}

// RecordUpdate displays information about an updated DNS resource record.
func (pp *TextPrinter) RecordUpdate(rr models.DNSRecord) {
	fmt.Printf("DNS resource record %s successfully updated\n", rr.Name)
}
