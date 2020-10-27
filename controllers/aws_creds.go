package controllers

import (
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
)

const (
	defaultAmazonWebservicesURN = "urn:amazon:webservices"
	defaultProfile              = "saml"
)

func formatAccount(url, login, role string, duration int) *cfg.IDPAccount {
	return &cfg.IDPAccount{
		URL:                  url,
		Username:             login,
		MFA:                  "Azure",
		Provider:             "ADFS",
		SkipVerify:           false,
		RoleARN:              role,
		SessionDuration:      duration,
		Profile:              defaultProfile,
		AmazonWebservicesURN: defaultAmazonWebservicesURN,
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

func getCredentials(assertion, role string, account *cfg.IDPAccount) (*awsconfig.AWSCredentials, error) {
	data, err := b64.StdEncoding.DecodeString(assertion)
	if err != nil {
		log.Error(err, "error decoding saml assertion")
		return nil, err
	}

	roles, err := saml2aws.ExtractAwsRoles(data)
	if err != nil {
		log.Error(err, "error parsing aws roles")
		return nil, err
	}

	awsRoles, err := saml2aws.ParseAWSRoles(roles)
	if err != nil {
		log.Error(err, "error parsing aws roles")
		return nil, err
	}

	awsRole, err := saml2aws.LocateRole(awsRoles, role)
	if err != nil {
		log.Error(err, "error locating role")
		return nil, err
	}

	awsCreds, err := loginToStsUsingRole(account, awsRole, assertion)
	if err != nil {
		log.Error(err, "error logging into aws role using saml assertion")
		return nil, err
	}
	return awsCreds, nil
}

func (r *Saml2AwsReconciler) createAWSCreds(saml *samletv1.Saml2Aws) (*awsconfig.AWSCredentials, string, error) {
	loginSecret, _ := r.readSecret(saml.Spec.SecretName, saml.Namespace)
	user, password := getLoginData(loginSecret)

	account := formatAccount(r.Config.IDPEndpoint, user, saml.Spec.RoleARN, r.Config.SessionDuration)
	provider, _ := adfs.New(account)
	loginDetails := &creds.LoginDetails{
		Username: account.Username,
		URL:      account.URL,
		Password: password,
	}
	samlAssertion, err := provider.Authenticate(loginDetails)
	if err != nil {
		log.Error(err, "error authenticating to IdP")
		return nil, "", err

	}
	awsCreds, err := getCredentials(samlAssertion, account.RoleARN, account)
	if err != nil {
		log.Error(err, "error logging into aws role using saml assertion")
		return nil, "", err
	}
	return awsCreds, account.Profile, nil
}
