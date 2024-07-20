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

package zones

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

// ListZones return lists zones on an account.
func (c *client) ListZones(ctx context.Context) ([]cloudflare.Zone, error) {
	zones, err := c.api.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	return zones, nil
}
