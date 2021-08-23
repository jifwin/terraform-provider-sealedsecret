package main

import (
	"github.com/akselleirv/terraform-sealed-secrets/sealed_secrets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return sealed_secrets.Provider()
		},
	})

}
