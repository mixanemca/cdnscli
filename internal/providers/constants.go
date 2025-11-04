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

package providers

// Provider type constants
const (
	TypeCloudflare = "cloudflare"
)

// DefaultDisplayNames contains default display names for provider types.
var DefaultDisplayNames = map[string]string{
	TypeCloudflare: "Cloudflare",
}

// GetDisplayName returns the display name for a provider type.
// If a custom display name is provided, it takes precedence.
func GetDisplayName(providerType string, customDisplayName string) string {
	if customDisplayName != "" {
		return customDisplayName
	}
	if displayName, ok := DefaultDisplayNames[providerType]; ok {
		return displayName
	}
	// Fallback to provider type if no default display name is found
	return providerType
}

