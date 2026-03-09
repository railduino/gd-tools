package dns

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

const (
	CloudflareBaseURL = " https://api.cloudflare.com/client/v4/zones"
)

type CloudflareProvider struct {
	AuthAPIToken string
	ApiClient    *cloudflare.Client
}

func Reference(token string) error {
	ctx := context.TODO()

	client := cloudflare.NewClient(option.WithAPIToken(token))
	zone, err := client.Zones.New(ctx, zones.ZoneNewParams{
		Account: cloudflare.F(zones.ZoneNewParamsAccount{
			ID: cloudflare.F("023e105f4ecef8ad9ca31a8372d0c353"),
		}),
		Name: cloudflare.F("example.com"),
		Type: cloudflare.F(zones.TypeFull),
	})
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", zone.ID)

	return nil
}
