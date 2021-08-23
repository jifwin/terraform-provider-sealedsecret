package kubeseal

import (
	"crypto/rsa"
	"fmt"
	"github.com/akselleirv/terraform-sealed-secrets/k8s"
	"k8s.io/client-go/util/cert"
)

type Kubeseal struct {
	c k8s.Clienter
}

func NewKubeseal(c k8s.Clienter) *Kubeseal {
	return &Kubeseal{c: c}
}

func (k *Kubeseal) FetchPK(controllerNamespace, controllerName string) (*rsa.PublicKey, error) {
	resp, err := k.c.Get(controllerNamespace, controllerName, "/v1/cert.pem")
	if err != nil {
		return nil, err
	}
	certs, err := cert.ParseCertsPEM(resp)
	if err != nil {
		return nil, err
	}
	pk, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected public key, got: %v", certs[0].PublicKey)
	}
	return pk, nil

}
