package sealedsecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/akselleirv/sealedsecret/kubeseal"
	"os"

	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	Name      = "name"
	Namespace = "namespace"
	Type      = "type"
	Secrets   = "secrets"
)

func resourceInGit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Schema: map[string]*schema.Schema{
			Name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the secret, must be unique",
			},
			Namespace: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Namespace of the secret",
			},
			Type: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Opaque",
				Description: "The secret type (ex. Opaque)",
			},
			Secrets: {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Key/value pairs to populate the secret",
			},
		},
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	secret, err := k8s.CreateSecret(
		d.Get(Name).(string),
		d.Get(Namespace).(string),
		d.Get(Type).(string),
		b64EncodeMapValue(d.Get(Secrets).(map[string]interface{})),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	cfg := m.(ProviderConfig)
	pk, err := kubeseal.FetchPK(cfg.Client, cfg.ControllerName, cfg.ControllerNamespace)
	if err != nil {
		return diag.FromErr(err)
	}

	sealedSecret, err := kubeseal.SealSecret(secret, pk)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := os.WriteFile("sealedSecret.yaml", sealedSecret, 0666); err != nil {
		return diag.FromErr(err)
	}

	return diag.Errorf("resource create ===========>")
}
func resourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("resource read ===========>")
}
func resourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("resource update ===========>")
}
func resourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("resource delete ===========>")
}

func b64EncodeMapValue(m map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		result[key] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", value)))
	}
	return result
}
