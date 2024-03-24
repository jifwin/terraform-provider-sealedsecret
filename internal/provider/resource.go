package provider

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/akselleirv/sealedsecret/internal/k8s"
	"github.com/akselleirv/sealedsecret/internal/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"log"
	"time"
)

const (
	name       = "name"
	namespace  = "namespace"
	secretType = "type"
	data       = "data"
	stringData = "string_data"
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
		CreateContext: resourceCreateLocal,
		ReadContext:   resourceCreateLocal,
		UpdateContext: resourceCreateLocal,
		DeleteContext: resourceDeleteLocal,
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
				Description: "The secret type (ex. Opaque). Default type is Opaque.",
			},
			data: {
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Description: "Key/value pairs to populate the secret. The value will be base64 encoded",
			},
			"yaml_content": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The produced sealed secret yaml file.",
			},
		},
	}
}

func resourceCreateLocal(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	provider := meta.(*ProviderConfig)
	filePath := d.Get(name).(string)

	logDebug("Creating sealed secret for path " + filePath)
	sealedSecret, err := createSealedSecret(ctx, provider, d)
	if err != nil {
		return diag.FromErr(err)
	}
	logDebug("Successfully created sealed secret for path " + filePath)

	d.SetId(filePath)
	d.Set(data, d.Get(data).(map[string]interface{}))
	d.Set("yaml_content", string(sealedSecret))

	return nil
}

func resourceDeleteLocal(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	var pk *rsa.PublicKey
	err = resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		var err error
		logDebug("Trying to fetch the public key")
		pk, err = provider.PublicKeyResolver(ctx)
		if err != nil {
			if k8sErrors.IsNotFound(err) || k8sErrors.IsServiceUnavailable(err) {
				logDebug("Retrying to fetch the public key: " + err.Error())
				return resource.RetryableError(fmt.Errorf("waiting for sealed-secret-controller to be deployed: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		logDebug("Successfully fetched the public key")
		return nil
	})

	if err != nil {
		return nil, err
	}

	return kubeseal.SealSecret(secret, pk)
}

func logDebug(msg string) {
	log.Printf("[DEBUG] %s\n", msg)
}
