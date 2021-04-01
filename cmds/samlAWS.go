package cmds

import (
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/willbtlr/imperson8/bash"
	"github.com/willbtlr/imperson8/saml"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	awsRegion          = "AWS_DEFAULT_REGION"
	awsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	awsSessionToken    = "AWS_SESSION_TOKEN"
)

type environ []string

func (e *environ) Unset(key string) {
	for i := range *e {
		if strings.HasPrefix((*e)[i], key+"=") {
			(*e)[i] = (*e)[len(*e)-1]
			*e = (*e)[:len(*e)-1]
			break
		}
	}
}

func (e *environ) Set(key, val string) {
	e.Unset(key)
	*e = append(*e, key+"="+val)
}

func SamlAWS(ctx *cli.Context) error {
	e := bash.CheckDeps("xmlsec1")
	if e != nil {
		return e
	}

	keyFile := ctx.String("key")
	responseFile := ctx.String("resp")
	principalARN := ctx.String("principal")
	roleARN := ctx.String("role")
	region := ctx.String("region")
	user := ctx.String("user")
	duration := ctx.Duration("duration")
	issuer := ctx.String("issuer")

	keyBs, e := ioutil.ReadFile(keyFile)
	if e != nil {
		return e
	}

	responseTemplateBs, e := ioutil.ReadFile(responseFile)
	if e != nil {
		return e
	}

	response, e := saml.SignResponse(keyBs, responseTemplateBs, saml.SignArgs{
		User: user,
		Issuer: issuer,
		Groups: []string{fmt.Sprintf("%s,%s", roleARN, principalARN)},
	})
	if e != nil {
		return e
	}

	encodedResponse := base64.StdEncoding.EncodeToString(response)

	sess, e := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if e != nil {
		return e
	}

	stsClient := sts.New(sess)
	stsResp, e := stsClient.AssumeRoleWithSAML(&sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int64(int64(duration.Seconds())),
		PrincipalArn:    aws.String(principalARN),
		RoleArn:         aws.String(roleARN),
		SAMLAssertion:   aws.String(encodedResponse),
	})
	if e != nil {
		return e
	}

	commandExec := ctx.Args().Slice()

	env := environ(os.Environ())
	env.Set(awsRegion, region)
	env.Set(awsAccessKeyID, *stsResp.Credentials.AccessKeyId)
	env.Set(awsSecretAccessKey, *stsResp.Credentials.SecretAccessKey)
	env.Set(awsSessionToken, *stsResp.Credentials.SessionToken)

	binary, e := exec.LookPath(commandExec[0])
	if e != nil {
		return e
	}

	return syscall.Exec(binary, commandExec, env)
}
