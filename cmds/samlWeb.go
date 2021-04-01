package cmds

import (
	"fmt"
	"github.com/willbtlr/imperson8/bash"
	"github.com/willbtlr/imperson8/browser"
	"github.com/willbtlr/imperson8/saml"
	"github.com/willbtlr/imperson8/server"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"os"
	"strings"
)

func SamlWeb(ctx *cli.Context) error {
	e := bash.CheckDeps("xmlsec1")
	if e != nil {
		return e
	}

	keyFile := ctx.String("key")
	responseFile := ctx.String("resp")
	user := ctx.String("user")
	groups := ctx.StringSlice("group")
	port := ctx.Int("port")
	issuer := ctx.String("issuer")

	var keyBs []byte
	keyFile = strings.TrimSpace(keyFile)
	if keyFile == "-" {
		keyBs, e = ioutil.ReadAll(os.Stdin)
	} else {
		keyBs, e = ioutil.ReadFile(keyFile)
	}
	if e != nil {
		return e
	}

	srv := server.Server{Port: port}
	e = srv.Start()
	if e != nil {
		return e
	}

	samlResponseBs, e := ioutil.ReadFile(responseFile)
	if e != nil {
		return e
	}

	signedResponse, e := saml.SignResponse(keyBs, samlResponseBs, saml.SignArgs{
		User:   user,
		Issuer: issuer,
		Groups: groups,
	})
	if e != nil {
		return e
	}

	srv.SetResponse(signedResponse)

	token, e := srv.SetToken()
	if e != nil {
		return e
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, token)

	e = browser.Open(url)
	<-srv.Done
	return e
}
