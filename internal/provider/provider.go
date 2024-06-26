package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jifwin/terraform-provider-sealedsecret/internal/k8s"
	"github.com/jifwin/terraform-provider-sealedsecret/internal/kubeseal"
)

const (
	kubernetes           = "kubernetes"
	host                 = "host"
	clientCertificate    = "client_certificate"
	clientKey            = "client_key"
	token                = "token"
	clusterCaCertificate = "cluster_ca_certificate"
	controllerName       = "controller_name"
	controllerNamespace  = "controller_namespace"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			kubernetes: {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Kubernetes configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						host: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The hostname (in form of URI) of Kubernetes master.",
						},
						token: {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_TOKEN", ""),
							Description: "Token to authenticate an service account",
						},
						clientCertificate: {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_CERT_DATA", ""),
							Description: "PEM-encoded client certificate for TLS authentication.",
						},
						clientKey: {
							Type:        schema.TypeString,
							Optional:    true,
							DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_KEY_DATA", ""),
							Description: "PEM-encoded client certificate key for TLS authentication.",
						},
						clusterCaCertificate: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "PEM-encoded root certificates bundle for TLS authentication.",
						},
					},
				},
			},
			controllerName: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the sealed-secret-controller.",
				Default:     "sealed-data-controller",
			},
			controllerNamespace: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The namespace the controller is running in.",
				Default:     "kube-system",
			},
		},
		ConfigureContextFunc: configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"sealedsecret": resourceLocal(),
		},
	}
}

type ProviderConfig struct {
	ControllerName      string
	ControllerNamespace string
	PublicKeyResolver   kubeseal.PKResolverFunc
}

func configureProvider(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	k8sCfg, ok := getMapFromSchemaSet(rd, kubernetes)
	if !ok {
		return nil, diag.FromErr(errors.New("k8s configuration is required"))
	}

	c, err := k8s.NewClient(&k8s.Config{
		Host:          k8sCfg[host].(string),
		ClusterCACert: []byte(k8sCfg[clusterCaCertificate].(string)),
		ClientCert:    []byte(k8sCfg[clientCertificate].(string)),
		ClientKey:     []byte(k8sCfg[clientKey].(string)),
		Token:         k8sCfg[token].(string),
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	cName := rd.Get(controllerName).(string)
	cNs := rd.Get(controllerNamespace).(string)

	return &ProviderConfig{
		ControllerName:      cName,
		ControllerNamespace: cNs,
		PublicKeyResolver:   kubeseal.FetchPK(c, cName, cNs),
	}, nil
}

func getMapFromSchemaSet(rd *schema.ResourceData, key string) (map[string]interface{}, bool) {
	m, ok := rd.GetOk(key)
	if !ok {
		return nil, ok
	}
	return m.([]interface{})[0].(map[string]interface{}), ok
}
