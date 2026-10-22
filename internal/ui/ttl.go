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

package ui

import (
	"strconv"
	"strings"
)

// autoTTL is the TTL value Cloudflare uses for "Auto". Proxied records always
// have it: the API accepts any TTL for them but keeps Auto.
const autoTTL = 1

// autoTTLLabel is how autoTTL is shown in the tables and the record editor.
const autoTTLLabel = "Auto"

// formatTTL renders a TTL for the UI, showing Auto instead of the raw 1.
func formatTTL(ttl int) string {
	if ttl == autoTTL {
		return autoTTLLabel
	}
	return strconv.Itoa(ttl)
}

// parseTTL converts a UI TTL value back to seconds. "Auto" (any case) maps to
// autoTTL; an unparsable value yields 0, the same as the previous behaviour.
func parseTTL(s string) int {
	if strings.EqualFold(strings.TrimSpace(s), autoTTLLabel) {
		return autoTTL
	}
	ttl, _ := strconv.Atoi(strings.TrimSpace(s))
	return ttl
}
