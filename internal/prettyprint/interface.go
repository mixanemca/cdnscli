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

import "github.com/cloudflare/cloudflare-go"

// PrettyPrinter interface for printing in various formats.
type PrettyPrinter interface {
	// ZonesList prints list of DNS zones.
	ZonesList(zones []cloudflare.Zone)
	// RecordsList prints list of DNS resource records.
	RecordsList(rrset []cloudflare.DNSRecord)
	// RecordInfo displays information about a specified DNS resource record.
	RecordInfo(rr cloudflare.DNSRecord)
	// RecordAdd displays information about a new DNS resource record.
	RecordAdd(rr cloudflare.DNSRecord)
	// RecordDel displays information about a deleted DNS recource record.
	RecordDel(rr cloudflare.DNSRecord)
	// RecordUpdate displays information about an updated DNS resource record.
	RecordUpdate(rr cloudflare.DNSRecord)
}
