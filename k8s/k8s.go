package k8s

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"time"
)

type Client struct {
	RestClient *corev1.CoreV1Client
}

const timeout = 10 * time.Second

type Clienter interface {
	Get(controllerName, controllerNamespace, path string) ([]byte, error)
}

func NewClient(host string, clusterCACert, clientCert, clientKey []byte) (*Client, error) {
	cfg := &rest.Config{
		Timeout: timeout,
	}
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

func (c *Client) Get(controllerName, controllerNamespace, path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var resp io.ReadCloser
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var innerErr error
		resp, innerErr = c.RestClient.
			Services(controllerNamespace).
			ProxyGet("http", controllerName, "", path, nil).
			Stream(ctx)
		if innerErr != nil {
			return innerErr
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("request to k8s cluster failed: %w", err)
	}
	b, err := io.ReadAll(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from k8 clsuter: %w", err)
	}
	return b, nil
}
