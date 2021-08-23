package k8s

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"time"
)

type Client struct {
	RestClient *corev1.CoreV1Client
}

type Clienter interface {
	Get(controllerNamespace, controllerName, path string) ([]byte, error)
}

func NewClient(host string, clusterCACert, clientCert, clientKey []byte) (*Client, error) {
	cfg := &rest.Config{}
	cfg.Host = host
	cfg.CAData = clusterCACert
	cfg.CertData = clientCert
	cfg.KeyData = clientKey

	c, err := corev1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{RestClient: c}, nil
}

func (c *Client) Get(controllerNamespace, controllerName, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := c.RestClient.
		Services(controllerNamespace).
		ProxyGet("http", controllerName, "", path, nil).
		Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	b, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to read response: %w", err)
	}
	return b, nil
}
