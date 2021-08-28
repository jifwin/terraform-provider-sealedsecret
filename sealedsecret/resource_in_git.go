package sealedsecret

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/akselleirv/sealedsecret/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	name          = "name"
	namespace     = "namespace"
	secretType    = "type"
	secrets       = "secrets"
	username      = "username"
	token         = "token"
	url           = "url"
	filepath      = "filepath"
	publicKeyHash = "public_key_hash"
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
			publicKeyHash: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public key hashed to detect if the public key changes.",
			},
		},
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(ProviderConfig)
	filePath := d.Get(filepath).(string)

	sealedSecret, err := createSealedSecret(&provider, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = provider.Git.Push(ctx, sealedSecret, filePath)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(filePath)
	if err := d.Set(secrets, d.Get(secrets).(map[string]interface{})); err != nil {
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}
func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(ProviderConfig)

	f, err := provider.Git.GetFile(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ssInGit := &SealedSecret{}
	if err := yaml.Unmarshal(f, ssInGit); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(name, ssInGit.Spec.Template.Metadata.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(namespace, ssInGit.Spec.Template.Metadata.Namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(secretType, ssInGit.Spec.Template.Type); err != nil {
		return diag.FromErr(err)
	}

	newPkHash := hashPublicKey(provider.PK)
	oldPkHash, ok := d.State().Attributes[publicKeyHash]
	if ok && newPkHash != oldPkHash {
		// If the PK changed then we are forcing it to be recreated.
		// We do not require any clean up since the keys stored in Git will be overwritten when applied again.
		// An improvement could be so notify the user the reason for the recreate was the PK change.
		d.SetId("")
	}

	if err := d.Set(publicKeyHash, newPkHash); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceCreate(ctx, d, meta)
}
func resourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(meta.(ProviderConfig).Git.DeleteFile(ctx, d.Get(filepath).(string)))
}

func createSealedSecret(provider *ProviderConfig, d *schema.ResourceData) ([]byte, error) {
	secret, err := k8s.CreateSecret(&k8s.SecretManifest{
		Name:      d.Get(name).(string),
		Namespace: d.Get(namespace).(string),
		Type:      d.Get(secretType).(string),
		Secrets:   b64EncodeMapValue(d.Get(secrets).(map[string]interface{})),
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

// The public key is hashed since we want to force update the resource if the key changes.
// Hashing the key also saves us some space.
func hashPublicKey(pk *rsa.PublicKey) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%v%v", pk.N, pk.E))))
}
