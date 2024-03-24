package provider

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/akselleirv/sealedsecret/internal/k8s"
	"github.com/akselleirv/sealedsecret/internal/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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

// The public key is hashed since we want to force update the resource if the key changes.
// Hashing the key also saves us some space.
func hashPublicKey(pk *rsa.PublicKey) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(fmt.Sprintf("%v%v", pk.N, pk.E))))
}

func logDebug(msg string) {
	log.Printf("[DEBUG] %s\n", msg)
}
