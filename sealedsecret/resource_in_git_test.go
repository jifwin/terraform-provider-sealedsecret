package sealedsecret

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"testing"
)

func TestSealedsecretInGit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"sealedsecret": func() (*schema.Provider, error) {
				return Provider(), nil
			},
		},
	})
}
