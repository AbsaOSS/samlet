package controllers

import (
	"bytes"
	"github.com/versent/saml2aws/v2/pkg/awsconfig"
	ini "gopkg.in/ini.v1"
	"k8s.io/api/core/v1"
)

func generateCredentiasFile(profile string, creds *awsconfig.AWSCredentials, secret *v1.Secret) (*v1.Secret, error) {
	iniFile := ini.Empty()
	sec, err := iniFile.NewSection(profile)
	if err != nil {
		return nil, err
	}

	err = sec.ReflectFrom(creds)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	_, err = iniFile.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	secret.Data["credentials"] = buf.Bytes()

	return secret, err
}

func generateEnvVariables(creds *awsconfig.AWSCredentials, secret *v1.Secret) *v1.Secret {
	secret.Data["AWS_ACCESS_KEY_ID"]        = []byte(creds.AWSAccessKey)
	secret.Data["AWS_SECRET_ACCESS_KEY"]    = []byte(creds.AWSSecretKey)
	secret.Data["AWS_SESSION_TOKEN"]        = []byte(creds.AWSSessionToken)
	secret.Data["X_SECURITY_TOKEN_EXPIRES"] = []byte(creds.Expires.String())
	secret.Data["X_PRINCIPAL_ARN"]          = []byte(creds.PrincipalARN)
	secret.Data["AWS_DEFAULT_REGION"]       = []byte(creds.Region)

	return secret
}
