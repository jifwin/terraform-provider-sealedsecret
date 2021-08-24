package sealedsecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/akselleirv/sealedsecret/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	name       = "name"
	namespace  = "namespace"
	secretType = "type"
	secrets    = "secrets"
	username   = "username"
	token      = "token"
	url        = "url"
	filepath   = "filepath"
)

func resourceInGit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Schema: map[string]*schema.Schema{
			name: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of the secret, must be unique",
			},
			namespace: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "namespace of the secret",
			},
			secretType: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Opaque",
				Description: "The secret type (ex. Opaque)",
			},
			secrets: {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Key/value pairs to populate the secret",
			},
			filepath: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The filepath in the Git repository. Including the filename itself and extension",
			},
		},
	}
}

func resourceCreate(ctx context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	secret, err := k8s.CreateSecret(
		rd.Get(name).(string),
		rd.Get(namespace).(string),
		rd.Get(secretType).(string),
		b64EncodeMapValue(rd.Get(secrets).(map[string]interface{})),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	provider := m.(ProviderConfig)
	pk, err := kubeseal.FetchPK(provider.Client, provider.ControllerName, provider.ControllerNamespace)
	if err != nil {
		return diag.FromErr(err)
	}

	sealedSecret, err := kubeseal.SealSecret(secret, pk)
	if err != nil {
		return diag.FromErr(err)
	}
	
	err = provider.Git.Push(ctx, sealedSecret, rd.Get(filepath).(string))
	if err != nil {
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
