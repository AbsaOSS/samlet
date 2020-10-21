package controllers

import (
	"context"
	b64 "encoding/base64"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	samletv1 "github.com/bison-cloud-platform/samlet/api/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/versent/saml2aws/v2"
	"github.com/versent/saml2aws/v2/pkg/awsconfig"
	"github.com/versent/saml2aws/v2/pkg/cfg"
	"github.com/versent/saml2aws/v2/pkg/creds"
	"github.com/versent/saml2aws/v2/pkg/provider/adfs"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DefaultEndpoint             = "https://sts.absa.co.za"
	DefaultAmazonWebservicesURN = "urn:amazon:webservices"
	DefaultSessionDuration      = 3600
	DefaultProfile              = "saml"
)

func formatAccount(url, login, role string) *cfg.IDPAccount {
	return &cfg.IDPAccount{
		URL:                  url,
		Username:             login,
		MFA:                  "Azure",
		Provider:             "ADFS",
		SkipVerify:           false,
		RoleARN:              role,
		SessionDuration:      DefaultSessionDuration,
		Profile:              DefaultProfile,
		AmazonWebservicesURN: DefaultAmazonWebservicesURN,
	}
}

func loginToStsUsingRole(account *cfg.IDPAccount, role *saml2aws.AWSRole, samlAssertion string) (*awsconfig.AWSCredentials, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: &account.Region,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	svc := sts.New(sess)

	params := &sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    aws.String(role.PrincipalARN), // Required
		RoleArn:         aws.String(role.RoleARN),      // Required
		SAMLAssertion:   aws.String(samlAssertion),     // Required
		DurationSeconds: aws.Int64(int64(account.SessionDuration)),
	}

	logrus.Infof("Requesting AWS credentials using SAML assertion")

	resp, err := svc.AssumeRoleWithSAML(params)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving STS credentials using SAML")
	}

	return &awsconfig.AWSCredentials{
		AWSAccessKey:     aws.StringValue(resp.Credentials.AccessKeyId),
		AWSSecretKey:     aws.StringValue(resp.Credentials.SecretAccessKey),
		AWSSessionToken:  aws.StringValue(resp.Credentials.SessionToken),
		AWSSecurityToken: aws.StringValue(resp.Credentials.SessionToken),
		PrincipalARN:     aws.StringValue(resp.AssumedRoleUser.Arn),
		Expires:          resp.Credentials.Expiration.Local(),
		Region:           account.Region,
	}, nil
}

func (r *Saml2AwsReconciler) createAWSCreds(request reconcile.Request, saml *samletv1.Saml2Aws) (*reconcile.Result, error) {
	loginSecret, _ := r.readSecret(saml.Spec.SecretName, saml.Namespace)
	user, password := getLoginData(loginSecret)

	if saml.Spec.IDPEndpoint == "" {
		saml.Spec.IDPEndpoint = DefaultEndpoint
	}
	account := formatAccount(saml.Spec.IDPEndpoint, user, saml.Spec.RoleARN)
	provider, _ := adfs.New(account)
	loginDetails := &creds.LoginDetails{
		Username: account.Username,
		URL:      account.URL,
		Password: password,
	}
	samlAssertion, err := provider.Authenticate(loginDetails)
	if err != nil {
		log.Error(err, "error authenticating to IdP")
		return &reconcile.Result{}, nil

	}

	data, err := b64.StdEncoding.DecodeString(samlAssertion)
	if err != nil {
		log.Error(err, "error decoding saml assertion")
		return &reconcile.Result{}, nil
	}

	roles, err := saml2aws.ExtractAwsRoles(data)
	if err != nil {
		log.Error(err, "error parsing aws roles")
		return &reconcile.Result{}, nil
	}

	awsRoles, err := saml2aws.ParseAWSRoles(roles)
	if err != nil {
		log.Error(err, "error parsing aws roles")
		return &reconcile.Result{}, nil
	}

	role, err := saml2aws.LocateRole(awsRoles, account.RoleARN)
	if err != nil {
		log.Error(err, "error locating role")
		return &reconcile.Result{}, nil
	}

	awsCreds, err := loginToStsUsingRole(account, role, samlAssertion)
	if err != nil {
		log.Error(err, "error logging into aws role using saml assertion")
		return &reconcile.Result{}, nil
	}
	iniData := generateIni(account.Profile, awsCreds)

	secret, _ := r.targetSecret(saml, iniData)
	err = r.Get(context.TODO(), types.NamespacedName{
		Namespace: saml.Namespace,
		Name:      saml.Spec.TargetSecretName,
	}, secret)

	if err != nil && k8s_errors.IsNotFound(err) {
		log.Info("Creating aws secret file", "Namespace", saml.Namespace, "Secret", saml.Spec.TargetSecretName)
		err = r.Create(context.TODO(), secret)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create secret", "Namespace", saml.Namespace, "Name", secret.Name)
			return &reconcile.Result{}, err
		}
	}

	return nil, nil
}
