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

package app

import (
	"github.com/mixanemca/cdnscli/internal/config"
	"github.com/mixanemca/cdnscli/internal/prettyprint"
)

// WithOutputFormat sets an app's output format
func WithOutputFormat(output prettyprint.OutputFormat) Option {
	return func(a *app) error {
		a.output = output
		return nil
	}
}

// WithConfig sets the application configuration
func WithConfig(cfg *config.Config) Option {
	return func(a *app) error {
		a.cfg = cfg
		return nil
	}
}

// WithProvider sets the provider name to use
func WithProvider(providerName string) Option {
	return func(a *app) error {
		a.providerName = providerName
		return nil
	}
}
