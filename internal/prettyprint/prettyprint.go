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

const (
	// FormatText format for human-readable output.
	FormatText OutputFormat = "text"
	// FormatJSON format for output in JSON.
	FormatJSON OutputFormat = "json"
	// FormatNone format for discarding output.
	FormatNone OutputFormat = "none"
)

// OutputFormat holds supported output formats.
type OutputFormat string

// New constructs a new TextPrinter
func New(output OutputFormat) PrettyPrinter {
	switch output {
	case FormatText:
		return &TextPrinter{}
	case FormatJSON:
		return &JSONPrinter{}
	case FormatNone:
		return &NonePrinter{}
	}

	// TODO: need to add validation of output-format flag
	// or change to enum flags
	return &TextPrinter{}
}
