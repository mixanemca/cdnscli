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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatTTL(t *testing.T) {
	assert.Equal(t, "Auto", formatTTL(1))
	assert.Equal(t, "300", formatTTL(300))
	assert.Equal(t, "0", formatTTL(0))
}

func TestParseTTL(t *testing.T) {
	assert.Equal(t, 1, parseTTL("Auto"))
	assert.Equal(t, 1, parseTTL("auto"))
	assert.Equal(t, 1, parseTTL(" AUTO "))
	assert.Equal(t, 300, parseTTL("300"))
	assert.Equal(t, 0, parseTTL("not-a-number"))
}

func TestFormatParseTTLRoundTrip(t *testing.T) {
	for _, ttl := range []int{1, 60, 300, 3600, 86400} {
		assert.Equal(t, ttl, parseTTL(formatTTL(ttl)))
	}
}
