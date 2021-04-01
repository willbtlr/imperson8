# Imperson8 

## Disclaimer

This is a security testing tool. Only use this on systems you have explicit authorization to test.

This isn't an exploit and won't help you get any access you don't already have. It's just a high level abstraction that makes it easier to use the access you have.

## The problem this is trying to solve

- Many systems delegate authentication and identity management to another system
- Examples include
    - mTLS
    - SSHCA
    - SAML
    - Kerberos
- On red team operations, I frequently run into these technologies
- If you get access to the identity provider's signing key, you can impersonate any user or group
- But there are many implementation specific subtleties involved when using these signing keys
- So I wrote this tool to provide a higher level abstraction allowing you to use these keys for security testing 
- Given the key, it makes it easier to do things like "log into the AWS management console assuming role X" without having to worry about the SAML mechanics

## Supported Delegated Auth Technologies
- Kubernetes mTLS client certificates
- SSH CA signed certificates
- SAML w/ web applications
- SAML w/ AWS SDK compliant CLI applications 

## OS Support

Tested on MacOS Catalina / Big Sur and Ubuntu 20.04.

## Installation

Make sure the following programs are on your $PATH:
- kubectl 
- xmlsec1
- ssh-add
- ssh-keygen

Compile:
```
$ go build
```

## Examples

### SAML Web

`imperson8 saml web` generates a signed SAML response and launches a browser impersonating the specified user / groups in a web application
```
imperson8 saml web \ 
  --key ./privkey.pem \
  --resp templates/aws.xml
  --issuer http://localhost:8080/auth/realms/master \
  --user wat \
  --group arn:aws:iam::9999999999:saml-provider/keycloak,arn:aws:iam::9999999999:role/ec2-admin
```

### SAML AWS

`imperson8 saml aws` uses `sts:AssumeRoleWithSAML` to get a set of temporary creds and then launches the specified program with that role configured
```
imperson8 saml aws \
  --user 'doesnt.exist@nonsense.com' \
  --key ./privkey.pem \
  --issuer 'http://localhost:8080/auth/realms/master' \
  --principal 'arn:aws:iam::99999999:saml-provider/keycloak' \
  --role 'arn:aws:iam::99999999:role/ec2-reader' \ 
  --region us-west-2 \
  --duration 900s \
  aws ec2 describe-instances
```

### k8s

```
imperson8 k8s \ 
  --user wbutler \
  --group system:masters \
  --key ca.key \
  --cert ca.crt
```

## Usage

### SAML Web
```
NAME:
   imperson8 saml web - launch a browser and send a signed challenge response to the SP URL

USAGE:
   imperson8 saml web [command options] [arguments...]

OPTIONS:
   --user value    user to impersonate
   --group value   list of groups to impersonate (pass flag multiple times for multiple groups)
   --key value     key file (default: "./key.pem")
   --resp value    path to SAML response xml file (default: "templates/aws.xml")
   --issuer value  Issuer metadata URL
   --port value    port for the local web server to listen on (default: 8443)
   --help, -h      show help (default: false)
   
```

### SAML AWS SDK

```
NAME:
   imperson8 saml aws - use AWS SDK compatible CLI tools with STS:AssumeRoleWithSAML

USAGE:
   imperson8 saml aws [command options] [arguments...]

OPTIONS:
   --user value       user to impersonate
   --key value        key file (default: "./key.pem")
   --resp value       path to SAML response xml file (default: "templates/aws.xml")
   --duration value   session TTL in golang duration string format (default: 2h0m0s)
   --issuer value     Issuer metadata URL
   --principal value  principal ARN for AssumeRoleWithSAML (usually the ARN of the SAML provider)
   --role value       ARN for the role you want to assume
   --region value     default region for AWS CLI (default: "us-east-1")
   --help, -h         show help (default: false)
```

### Kubernetes

```
NAME:
   imperson8 k8s - add an auth method to .kube/config using a signed client cert impersonating the specified user / group

USAGE:
   imperson8 k8s [command options] [arguments...]

OPTIONS:
   --user value      user to impersonate
   --group value     list of groups to impersonate (pass flag multiple times for multiple groups)
   --key value       key file (default: "./key.pem")
   --duration value  session TTL in golang duration string format (default: 2h0m0s)
   --cert value      CA cert file (default: "cert.pem")
   --help, -h        show help (default: false)
```

### SSHCA

```
NAME:
   imperson8 sshca - generate a signed SSHCA certificate and private key impersonating the speficied user / group

USAGE:
   imperson8 sshca [command options] [arguments...]

OPTIONS:
   --user value      user to impersonate
   --group value     list of groups to impersonate (pass flag multiple times for multiple groups)
   --key value       key file (default: "./key.pem")
   --duration value  session TTL in golang duration string format (default: 2h0m0s)
   --out value       base key filename (default: "user_key")
   --help, -h        show help (default: false)
```
