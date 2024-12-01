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
	"context"
	"os"

	"github.com/cloudflare/cloudflare-go"
	pp "github.com/mixanemca/cfdnscli/internal/prettyprint"
	"github.com/mixanemca/cfdnscli/internal/zones"
)

type app struct {
	api    *cloudflare.API
	zones  zones.Client
	pp     pp.PrettyPrinter
	output pp.OutputFormat
}

// Option options for app
type Option func(c *app) error

// New creates a new CloudFlare API client. Various client options can be used to configure
// the CloudFlare client
func New(opts ...Option) (App, error) {
	// App with default values
	a := &app{}

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	// TODO: get the token from config
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		return nil, err
	}

	// Verify token
	_, err = api.VerifyAPIToken(context.Background())
	if err != nil {
		return nil, err
	}

	a.api = api

	a.zones = zones.New(a.api)

	a.pp = pp.New(pp.OutputFormat(a.output))

	return a, nil
}

func (a *app) Zones() zones.Client {
	return a.zones
}

func (a *app) Printer() pp.PrettyPrinter {
	return a.pp
}
