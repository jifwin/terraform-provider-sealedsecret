package k8s

import (
	"bytes"
	"errors"

	"html/template"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

const secretManifestTmpl = `
apiVersion: v1
kind: Secret
metadata:
  creationTimestamp: null
  name: {{ .Name }}
  namespace: {{ .Namespace }}
data:
  {{- range $key, $value := .Secrets }}
  {{ $key }}: {{ $value -}}
  {{ end }}
type: {{ .Type }}`

type SecretManifest struct {
	Name      string
	Namespace string
	Type      string
	Secrets   map[string]interface{}
}

func CreateSecret(sm *SecretManifest) (v1.Secret, error) {
	secretManifestYAML := new(bytes.Buffer)

	t, err := template.New("secretManifestTmpl").Parse(secretManifestTmpl)
	if err != nil {
		return v1.Secret{}, err
	}
	if err := t.Execute(secretManifestYAML, sm); err != nil {
		return v1.Secret{}, err
	}

	var secret v1.Secret
	if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), secretManifestYAML.Bytes(), &secret); err != nil {
		return v1.Secret{}, err
	}

	if len(secret.Data) == 0 && len(secret.StringData) == 0 {
		return v1.Secret{}, errors.New("unable to create a secret with empty Data and StringData ")
	}

	return secret, nil
}
