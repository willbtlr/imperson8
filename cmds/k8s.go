package cmds

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/willbtlr/imperson8/bash"
	"github.com/urfave/cli/v2"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func K8S(ctx *cli.Context) error {
	e := bash.CheckDeps("kubectl")
	if e != nil {
		return e
	}

	homeDir := os.Getenv("HOME")
	kubeDir := filepath.Join(homeDir, ".kube")
	clientCertFileName := filepath.Join(kubeDir, "client_cert.pem")
	clientKeyFileName := filepath.Join(kubeDir, "client_key.pem")

	user := ctx.String("user")
	groups := ctx.StringSlice("group")
	caCertFile := ctx.String("cert")
	caKeyFile := ctx.String("key")
	duration := ctx.Duration("duration")

	// parse CA certificates
	caCerts, e := tls.LoadX509KeyPair(caCertFile, caKeyFile)
	if e != nil {
		return e
	}

	caCert, e := x509.ParseCertificate(caCerts.Certificate[0])
	if e != nil {
		return e
	}

	// generate and sign a client certificate
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		return e
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, e := rand.Int(rand.Reader, serialNumberLimit)
	if e != nil {
		return e
	}

	csrTmpl := &x509.Certificate{
		Subject:            pkix.Name{CommonName: user, Organization: groups},
		IsCA:               false,
		SerialNumber:       serialNumber,
		NotBefore:          time.Now().Add(-5 * time.Minute),
		NotAfter:           time.Now().Add(duration),
		SignatureAlgorithm: x509.SHA256WithRSA,
		KeyUsage:           x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}

	certificateBs, e := x509.CreateCertificate(rand.Reader, csrTmpl, caCert, key.Public().(*rsa.PublicKey), caCerts.PrivateKey)
	if e != nil {
		return e
	}

	// save client certificate to disk to include in kube config
	clientCertFile, e := os.OpenFile(clientCertFileName, os.O_WRONLY|os.O_CREATE, 0700)
	if e != nil {
		return e
	}
	defer clientCertFile.Close()

	clientKeyFile, e := os.OpenFile(clientKeyFileName, os.O_WRONLY|os.O_CREATE, 0700)
	if e != nil {
		return e
	}
	defer clientKeyFile.Close()

	e = pem.Encode(clientCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: certificateBs})
	if e != nil {
		return e
	}

	privateKeyBs, e := x509.MarshalPKCS8PrivateKey(key)
	if e != nil {
		return e
	}

	e = pem.Encode(clientKeyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyBs})
	if e != nil {
		return e
	}

	// add credentials to kube config and associate them with current context
	_, e = bash.Execf(`kubectl config set-credentials %s --client-certificate=%s --client-key=%s`, user, clientCertFileName, clientKeyFileName)
	if e != nil {
		return e
	}

	_, e = bash.Execf(`kubectl config set-context --current --user=%s`, user)
	return e
}
