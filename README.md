# samlet
Samlet is a Kubernetes operator based on [saml2aws](https://github.com/Versent/saml2aws)

# Why
Samlet allows you to generate Kubernetes secrets in either `envVariables` or `credentialsFile` and rotate them 10 minutes before expiration period. Reloading 
credentials and watching expiration time logic is left to a consumer.

# Example
## envVariables
Following `Saml2Aws` resource will request AWS credentials valid for 2 hours using `example-login` (credentail keys are `username` and `password`) credentials and create new `target-secret`.
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
## credentialsFile
Credentials file type formats target secret content in a way so it can be mounted into a Pod as a volume. Once mounted it can be used as standard `.aws/credentials` ini file.
