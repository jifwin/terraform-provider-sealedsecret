package sealedsecret

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/akselleirv/sealedsecret/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
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

type SealedSecret struct {
	Spec struct {
		EncryptedData map[string]string `yaml:"encryptedData"`
		Template      struct {
			Type     string `yaml:"type"`
			Metadata struct {
				Name      string `yaml:"name"`
				Namespace string `yaml:"namespace"`
			} `yaml:"metadata"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

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
				Sensitive:   true,
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
	provider := m.(ProviderConfig)
	filePath := rd.Get(filepath).(string)

	sealedSecret, err := createSealedSecret(&provider, rd)
	if err != nil {
		return diag.FromErr(err)
	}

	err = provider.Git.Push(ctx, sealedSecret, filePath)
	if err != nil {
		return diag.FromErr(err)
	}
	rd.SetId(filePath)

	return resourceRead(ctx, rd, m)
}
func resourceRead(ctx context.Context, rd *schema.ResourceData, m interface{}) diag.Diagnostics {
	provider := m.(ProviderConfig)

	f, err := provider.Git.GetFile(rd.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ssInGit := &SealedSecret{}
	if err := yaml.Unmarshal(f, ssInGit); err != nil {
		return diag.FromErr(err)
	}

	if err := rd.Set(name, ssInGit.Spec.Template.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := rd.Set(namespace, ssInGit.Spec.Template.Metadata.Namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := rd.Set(secretType, ssInGit.Spec.Template.Type); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("resource update ===========>")
}
func resourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("resource delete ===========>")
}

func createSealedSecret(provider *ProviderConfig, rd *schema.ResourceData) ([]byte, error) {
	secret, err := k8s.CreateSecret(&k8s.SecretManifest{
		Name:      rd.Get(name).(string),
		Namespace: rd.Get(namespace).(string),
		Type:      rd.Get(secretType).(string),
		Secrets:   b64EncodeMapValue(rd.Get(secrets).(map[string]interface{})),
	})
	if err != nil {
		return nil, err
	}

	return kubeseal.SealSecret(secret, provider.PK)
}

func b64EncodeMapValue(m map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		result[key] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", value)))
	}
	return result
}

func handleEncryptedSecretsDiff(ctx context.Context, old, new, meta interface{}) bool {
	return true
}
