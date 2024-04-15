package k8s

import (
	"errors"
	v1 "k8s.io/api/core/v1"
)

type SecretManifest struct {
	Name      string
	Namespace string
	Type      string
	Data      map[string]string
}

var ErrEmptyData = errors.New("secret manifest Data and StringData cannot be empty")

func CreateSecret(sm *SecretManifest) (v1.Secret, error) {
	if len(sm.Data) == 0 {
		return v1.Secret{}, ErrEmptyData
	}

	data := make(map[string][]byte)
	for key, value := range sm.Data {
		data[key] = []byte(value)
	}

	var secret v1.Secret
	secret.APIVersion = "v1"
	secret.Kind = "Secret"
	secret.ObjectMeta.Name = sm.Name
	secret.ObjectMeta.Namespace = sm.Namespace
	secret.Data = data
	secret.Type = v1.SecretType(sm.Type)

	return secret, nil
}
