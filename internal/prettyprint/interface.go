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

// Package prettyprint holds pretty printer for internal models.
package prettyprint

import (
	"github.com/mixanemca/cdnscli/internal/models"
)

// PrettyPrinter interface for printing in various formats.
type PrettyPrinter interface {
	// ZonesList prints list of DNS zones.
	ZonesList(zones []models.Zone)
	// RecordsList prints list of DNS resource records.
	RecordsList(rrset []models.DNSRecord)
	// RecordInfo displays information about a specified DNS resource record.
	RecordInfo(rr models.DNSRecord)
	// RecordAdd displays information about a new DNS resource record.
	RecordAdd(rr models.DNSRecord)
	// RecordDel displays information about a deleted DNS recource record.
	RecordDel(rr models.DNSRecord)
	// RecordUpdate displays information about an updated DNS resource record.
	RecordUpdate(rr models.DNSRecord)
}
