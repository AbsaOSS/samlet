package controllers

import (
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	userKey = "username"
	passKey = "password"
)

func getLoginData(secret *v1.Secret) (string, string) {
	user := string(secret.Data[userKey])
	pass := string(secret.Data[passKey])
	return user, pass
}

func (r *Saml2AwsReconciler) readSecret(name, namespace string) (*v1.Secret, error) {
	loginSecret := &v1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, loginSecret)
	if err != nil {
		log.Error(err, "Failed to read secret")
		return nil, err
	}
	return loginSecret, nil
}
