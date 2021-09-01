package sealedsecret

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/akselleirv/sealedsecret/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
)

const (
	name          = "name"
	namespace     = "namespace"
	secretType    = "type"
	data          = "data"
	stringData    = "string_data"
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
			data: {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Description: "Key/value pairs to populate the secret. The value will be base64 encoded",
			},
			stringData: {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Description: "Key/value pairs to populate the secret.",
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
	if err := d.Set(data, d.Get(data).(map[string]interface{})); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(stringData, d.Get(stringData).(map[string]interface{})); err != nil {
		return diag.FromErr(err)
	}

	return resourceRead(ctx, d, meta)
}
func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(ProviderConfig)

	f, err := provider.Git.GetFile(d.Id())
	if errors.Is(err, os.ErrNotExist) {
		d.SetId("")
		return nil
	}

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

	pk, err := provider.PublicKeyResolver()
	if err != nil {
		return diag.FromErr(err)
	}
	newPkHash := hashPublicKey(pk)
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
	rawSecret := k8s.SecretManifest{
		Name:      d.Get(name).(string),
		Namespace: d.Get(namespace).(string),
		Type:      d.Get(secretType).(string),
	}
	if dataRaw, ok := d.GetOk(data); ok {
		rawSecret.Data = dataRaw.(map[string]interface{})
	}
	if stringDataRaw, ok := d.GetOk(stringData); ok {
		m := make(map[string]string)
		for k, v := range stringDataRaw.(map[string]interface{}) {
			m[k] = v.(string)
		}
		rawSecret.StringData = m
	}

	secret, err := k8s.CreateSecret(&rawSecret)
	if err != nil {
		return nil, err
	}

	pk, err := provider.PublicKeyResolver()
	if err != nil {
		return nil, err
	}

	return kubeseal.SealSecret(secret,pk)
}

// The public key is hashed since we want to force update the resource if the key changes.
// Hashing the key also saves us some space.
func hashPublicKey(pk *rsa.PublicKey) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%v%v", pk.N, pk.E))))
}
