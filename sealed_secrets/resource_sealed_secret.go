package sealed_secrets

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceSealedSecret() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the secret, must be unique",
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Namespace of the secret",
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The secret type (ex. Opaque)",
			},
			"secrets": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Key/value pairs to populate the secret",
			},
			"controller_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the SealedSecrets controller in the cluster",
			},
			"controller_namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Namespace of the SealedSecrets controller in the cluster",
			},
		},
	}
}
