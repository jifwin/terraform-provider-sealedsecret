package provider

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jifwin/terraform-provider-sealedsecret/internal/k8s"
	"github.com/jifwin/terraform-provider-sealedsecret/internal/kubeseal"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	name         = "name"
	namespace    = "namespace"
	secretType   = "type"
	data         = "data"
	yaml_content = "yaml_content"
	public_key   = "public_key"
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

func resourceLocal() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a sealed secret and store it in yaml_content.",
		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		DeleteContext: resourceDelete,
		Schema: map[string]*schema.Schema{
			name: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "name of the secret, must be unique",
			},
			namespace: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "namespace of the secret",
			},
			secretType: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Opaque",
				ForceNew:    true,
				Description: "The secret type (ex. Opaque). Default type is Opaque.",
			},
			data: {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "Key/value pairs to populate the secret. The value will be base64 encoded",
			},
			yaml_content: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The produced sealed secret yaml file.",
			},
			public_key: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The key used for encryption",
			},
		},
	}
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*ProviderConfig)
	pk, err := getPublicKey(ctx, provider)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges(name, namespace, secretType, data) || (formatPublicKeyAsString(pk) != d.Get("public_key")) {
		return resourceCreate(ctx, d, meta)
	} else {
		return nil
	}
}
func resourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*ProviderConfig)
	filePath := d.Get(name).(string)
	pk, err := getPublicKey(ctx, provider)
	if err != nil {
		return diag.FromErr(err)
	}

	logDebug("Creating sealed secret for path " + filePath)
	sealedSecret, err := createSealedSecret(ctx, provider, d)
	if err != nil {
		return diag.FromErr(err)
	}
	logDebug("Successfully created sealed secret for path " + filePath)

	d.SetId(filePath)
	d.Set(data, d.Get(data).(map[string]interface{})) //TODO: update
	d.Set("yaml_content", string(sealedSecret))
	d.Set("public_key", formatPublicKeyAsString(pk))

	return nil
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func createSealedSecret(ctx context.Context, provider *ProviderConfig, d *schema.ResourceData) ([]byte, error) {
	rawSecret := k8s.SecretManifest{
		Name:      d.Get(name).(string),
		Namespace: d.Get(namespace).(string),
		Type:      d.Get(secretType).(string),
	}
	if dataRaw, ok := d.GetOk(data); ok {
		data := make(map[string]string)
		for key, value := range dataRaw.(map[string]interface{}) {
			data[key] = value.(string)
		}
		rawSecret.Data = data
	}

	secret, err := k8s.CreateSecret(&rawSecret)
	if err != nil {
		return nil, err
	}

	pk, err := getPublicKey(ctx, provider)

	if err != nil {
		return nil, err
	}

	return kubeseal.SealSecret(secret, pk)
}

func getPublicKey(ctx context.Context, provider *ProviderConfig) (*rsa.PublicKey, error) {
	var pk *rsa.PublicKey
	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		var err error
		logDebug("Trying to fetch the public key")
		pk, err = provider.PublicKeyResolver(ctx)
		if err != nil {
			//TODO: check the actual type of error (e.g. timeout)
			if true || k8sErrors.IsNotFound(err) || k8sErrors.IsServiceUnavailable(err) {
				logDebug("Retrying to fetch the public key: " + err.Error())
				return resource.RetryableError(fmt.Errorf("waiting for sealed-secret-controller to be deployed: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		logDebug("Successfully fetched the public key")
		return nil
	})
	return pk, err
}

// TODO: refactor
func formatPublicKeyAsString(pk *rsa.PublicKey) string {
	return strings.Join([]string{pk.N.String(), strconv.Itoa(pk.E)}, "::")
}

func logDebug(msg string) {
	log.Printf("[DEBUG] %s\n", msg)
}
