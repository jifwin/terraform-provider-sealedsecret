package sealed_secrets

import (
	"bytes"
	"context"
	"github.com/akselleirv/terraform-sealed-secrets/k8s"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"kubernetes": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    false,
				Description: "Kubernetes configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_HOST", ""),
							Description: "The hostname (in form of URI) of Kubernetes master.",
						},
						"client_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_CERT_DATA", ""),
							Description: "PEM-encoded client certificate for TLS authentication.",
						},
						"client_key": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_KEY_DATA", ""),
							Description: "PEM-encoded client certificate key for TLS authentication.",
						},
						"cluster_ca_certificate": {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLUSTER_CA_CERT_DATA", ""),
							Description: "PEM-encoded root certificates bundle for TLS authentication.",
						},
					},
				},
			},
		},
		ConfigureContextFunc: configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"sealed_secret": resourceSealedSecret(),
		},
	}
}

func configureProvider(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var (
		host          string
		clusterCaCert []byte
		clientCert    []byte
		clientKey     []byte
	)

	if v, ok := rd.GetOk("host"); ok {
		host = v.(string)
	}
	if v, ok := rd.GetOk("cluster_ca_certificate"); ok {
		clusterCaCert = toBytes(v)
	}
	if v, ok := rd.GetOk("client_certificate"); ok {
		clientCert = toBytes(v)
	}
	if v, ok := rd.GetOk("client_key"); ok {
		clientKey = toBytes(v)
	}
	c, err := k8s.NewClient(host, clusterCaCert, clientCert, clientKey)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return c, nil
}

func toBytes(v interface{}) []byte {
	return bytes.NewBufferString(v.(string)).Bytes()
}
