package main

import (
	"github.com/akselleirv/sealedsecret/sealedsecret"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: sealedsecret.Provider,
	})

}
