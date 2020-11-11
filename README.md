# samlet
Samlet is a Kubernetes operator based on [saml2aws](https://github.com/Versent/saml2aws)

# Why
Samlet provides Kubernetes applications tied to organization IdP with seamless access to AWS resources via [SAML 2.0 identity federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_saml.html).  
It stores generated AWS session credentials as k8s Secrets, so that they can be consumed by application container as mounted [AWS credentials file](https://docs.aws.amazon.com/credref/latest/refdocs/creds-config-files.html) or wired as [AWS SDK environment variables](https://docs.aws.amazon.com/credref/latest/refdocs/environment-variables.html).
Secrets are automatically rotated 10 minutes before expiration period. Reloading 
credentials and watching expiration time logic is left to a consumer.

# Example
## Environment variables
Following `Saml2Aws` Custom Resource, once created in k8s cluster, will case Samlet operator to request AWS credentials valid for 2 hours using `example-login` (credential keys are `username` and `password`) credentials and create new `target-secret` with AWS SDK environment variables. 
These environment variables can be then wired from the secret to an application pod using `envFrom` option in Pod manifest.
```yaml
apiVersion: samlet.absa.oss/v1
kind: Saml2Aws
metadata:
  name: saml2aws-sample
spec:
  # Secret contains username/password to authenticate against IDP (type ADFS)
  secretName: examlpe-login
  # Secret to be created by controller containing issued AWS credentials
  targetSecretName: target-secret
  # Format for generated secret (env-file or ini)
  secretFormat: envVariables
  # Sepcify validitity time
  sessionDuration: 2h
  # The ARN of the role to assume
  roleARN: arn:aws:iam::888888888888:role/adfs-saml2aws-sample-role
```
resulting `envVariables` type secret could be consumed in Pod like:
```yaml
    envFrom:
      - secretRef:
         name: target-secret
```
## Credentials file
Credentials file type formats target secret content in a way so it can be mounted into a Pod as a volume. Once mounted it can be used as standard `.aws/credentials` ini file.
