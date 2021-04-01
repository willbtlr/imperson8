package main

import (
	"github.com/willbtlr/imperson8/cmds"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var (
	userFlag = &cli.StringFlag{
		Name:     "user",
		Usage:    "user to impersonate",
		Required: true,
	}
	groupsFlag = &cli.StringSliceFlag{
		Name:     "group",
		Usage:    "list of groups to impersonate (pass flag multiple times for multiple groups)",
		Required: true,
	}
	keyFlag = &cli.StringFlag{
		Name:  "key",
		Usage: "key file",
		Value: "./key.pem",
	}
	respFlag = &cli.StringFlag{
		Name:  "resp",
		Usage: "path to SAML response xml file",
		Value: "templates/aws.xml",
	}
	durationFlag = &cli.DurationFlag{
		Name:  "duration",
		Usage: "session TTL in golang duration string format",
		Value: 2 * time.Hour,
	}
	issuerFlag = &cli.StringFlag{
		Name:  "issuer",
		Usage: "Issuer metadata URL",
	}
)

func main() {
	app := &cli.App{
		Name:  "imperson8",
		Usage: "use signing keys to imperonate legitimate users and groups",
		Commands: []*cli.Command{
			{
				Name:  "saml",
				Usage: "impersonate SAML users",
				Subcommands: []*cli.Command{
					{
						Name:  "web",
						Usage: "launch a browser and send a signed challenge response to the ACS URL",
						Flags: []cli.Flag{
							userFlag,
							groupsFlag,
							keyFlag,
							respFlag,
							issuerFlag,
							&cli.IntFlag{
								Name:  "port",
								Usage: "port for the local web server to listen on",
								Value: 8443,
							},
						},
						Action: cmds.SamlWeb,
					},
					{
						Name:  "aws",
						Usage: "use AWS SDK compatible CLI tools with STS:AssumeRoleWithSAML",
						Flags: []cli.Flag{
							userFlag,
							keyFlag,
							respFlag,
							durationFlag,
							issuerFlag,
							&cli.StringFlag{
								Name:     "principal",
								Usage:    "principal ARN for AssumeRoleWithSAML (usually the ARN of the SAML provider)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "role",
								Usage:    "ARN for the role you want to assume",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "region",
								Usage: "default region for AWS CLI",
								Value: "us-east-1",
							},
						},
						Action: cmds.SamlAWS,
					},
				},
			},
			{
				Name:  "k8s",
				Usage: "add an auth method to .kube/config using a signed client cert impersonating the specified user / group",
				Flags: []cli.Flag{
					userFlag,
					groupsFlag,
					keyFlag,
					durationFlag,
					&cli.StringFlag{
						Name:  "cert",
						Usage: "CA cert file",
						Value: "cert.pem",
					},
				},
				Action: cmds.K8S,
			},
			{
				Name:  "sshca",
				Usage: "generate a signed SSHCA certificate and private key impersonating the speficied user / group",
				Flags: []cli.Flag{
					userFlag,
					groupsFlag,
					keyFlag,
					durationFlag,
					&cli.StringFlag{
						Name:  "out",
						Usage: "base key filename",
						Value: "user_key",
					},
				},
				Action: cmds.SSHCA,
			},
		},
	}

	e := app.Run(os.Args)
	if e != nil {
		logrus.WithError(e).Fatal("error starting app")
	}
}
