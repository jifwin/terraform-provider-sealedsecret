package provider

import (
	"context"
	"errors"
	"github.com/akselleirv/sealedsecret/internal/k8s"
	"github.com/akselleirv/sealedsecret/internal/kubeseal"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
							Required:    false,
							Description: "Bearer token for authentication",
						},
						clientCertificate: {
							Type:        schema.TypeString,
							Required:    false,
							Description: "PEM-encoded client certificate for TLS authentication.",
						},
						clientKey: {
							Type:        schema.TypeString,
							Required:    false,
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
	Client              *k8s.Client
	PublicKeyResolver   kubeseal.PKResolverFunc
}

// TODO: context not used?
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
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	cName := rd.Get(controllerName).(string)
	cNs := rd.Get(controllerNamespace).(string)

	return &ProviderConfig{
		ControllerName:      cName,
		ControllerNamespace: cNs,
		Client:              c,
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
