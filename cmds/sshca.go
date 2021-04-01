package cmds

import (
	"github.com/willbtlr/imperson8/bash"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
	"time"
)

func SSHCA(ctx *cli.Context) error {
	e := bash.CheckDeps("ssh-add", "ssh-keygen")
	if e != nil {
		return e
	}

	keyFile := ctx.String("key")
	user := ctx.String("user")
	groups := ctx.StringSlice("group")
	duration := ctx.Duration("duration")
	outFile := ctx.String("out")

	e = os.Remove(outFile)
	if e != nil && !os.IsNotExist(e) {
		return e
	}

	_, e = bash.Execf("ssh-keygen -f %s -N ''", outFile)
	if e != nil {
		return e
	}

	exp := "always:" + time.Now().Add(duration).Format("200601021504")
	_, e = bash.Execf(`ssh-keygen -s %s -I %s -n %s -V '%s' %s`, keyFile, user, strings.Join(groups, ","), exp, outFile)
	if e != nil {
		return e
	}

	_, e = bash.Execf("ssh-add -D")
	if e != nil {
		return e
	}

	_, e = bash.Execf("ssh-add %s", outFile)
	return e
}
