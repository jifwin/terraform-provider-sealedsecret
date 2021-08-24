package sealedsecret

import (
	"context"

	"github.com/akselleirv/sealedsecret/k8s"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	kubernetes           = "kubernetes"
	host                 = "host"
	clientCertificate    = "client_certificate"
	clientKey            = "client_key"
	clusterCaCertificate = "cluster_ca_certificate"
	controllerName       = "controller_name"
	controllerNamespace  = "controller_namespace"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			kubernetes: {
				Type:        schema.TypeSet,
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
						clientCertificate: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "PEM-encoded client certificate for TLS authentication.",
						},
						clientKey: {
							Type:        schema.TypeString,
							Required:    true,
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
				Default:     "sealed-secrets-controller",
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
			"sealedsecret_in_git": resourceInGit(),
		},
	}
}

type ProviderConfig struct {
	ControllerName      string
	ControllerNamespace string
	Client              *k8s.Client
}

func configureProvider(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// this is safe since the TypeSet is set to required and max is 1
	k8sCfg := rd.Get(kubernetes).(*schema.Set).List()[0].(map[string]interface{})

	c, err := k8s.NewClient(
		k8sCfg[host].(string),
		[]byte(k8sCfg[clusterCaCertificate].(string)),
		[]byte(k8sCfg[clientCertificate].(string)),
		[]byte(k8sCfg[clientKey].(string)),
	)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return ProviderConfig{
		ControllerName:      rd.Get(controllerName).(string),
		ControllerNamespace: rd.Get(controllerNamespace).(string),
		Client:              c,
	}, nil
}
