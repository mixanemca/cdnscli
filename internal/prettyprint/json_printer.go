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
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

// JSONPrinter prints in JSON format.
type JSONPrinter struct{}

// ZonesList prints list of DNS zones.
func (pp *JSONPrinter) ZonesList(zones []cloudflare.Zone) {
	fmt.Println(marshalJSON(zones))
}

// RecordsList prints list of DNS resource records.
func (pp *JSONPrinter) RecordsList(rrset []cloudflare.DNSRecord) {
	fmt.Println(marshalJSON(rrset))
}

// RecordInfo displays information about a specified DNS resource record.
func (pp *JSONPrinter) RecordInfo(rr cloudflare.DNSRecord) {
	fmt.Println(marshalJSON(rr))
}

// RecordAdd displays information about a new DNS resource record.
func (pp *JSONPrinter) RecordAdd(rr cloudflare.DNSRecord) {
	fmt.Println(marshalJSON(rr))
}

// RecordDel displays information about a deleted DNS recource record.
func (pp *JSONPrinter) RecordDel(rr cloudflare.DNSRecord) {
	fmt.Println(marshalJSON(rr))
}

// RecordUpdate displays information about an updated DNS resource record.
func (pp *JSONPrinter) RecordUpdate(rr cloudflare.DNSRecord) {
	fmt.Println(marshalJSON(rr))
}

func marshalJSON(v any) string {
	j, _ := json.Marshal(v)
	return string(j)
}
