package k8s

import (
	"encoding/base64"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type SecretManifest struct {
	Name      string
	Namespace string
	Type      string
	Data      map[string]string
	//Deprecated: TODO: remove
	StringData map[string]string
}

var ErrEmptyData = errors.New("secret manifest Data and StringData cannot be empty")

func CreateSecret(sm *SecretManifest) (v1.Secret, error) {
	if len(sm.Data) == 0 {
		return v1.Secret{}, ErrEmptyData
	}

	//TODO: rethink
	// if it is a .docker/config.json file then the data should already be base64 encoded
	//if sm.Type != "kubernetes.io/dockerconfigjson" {
	//	sm.Data = b64EncodeMapValue(sm.Data)
	//}

	data := make(map[string][]byte)
	for key, value := range sm.Data {
		data[key] = []byte(value)
	}

	var secret v1.Secret
	secret.APIVersion = "v1"
	secret.Kind = "Secret" //TODO: is it really needed?
	secret.ObjectMeta.Name = sm.Name
	secret.ObjectMeta.Namespace = sm.Namespace
	secret.Data = data
	secret.Type = v1.SecretType(sm.Type)

	return secret, nil
}

// TODO: to be removed?
func b64EncodeMapValue(m map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		result[key] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", value)))
	}
	return result
}
