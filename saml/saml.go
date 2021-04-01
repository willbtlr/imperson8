package saml

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"github.com/edaniels/go-saml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"
)

func Destination(samlAssertionBs []byte) (string, error) {
	res := saml.Response{}
	e := xml.Unmarshal(samlAssertionBs, &res)
	if e != nil {
		return "", e
	}

	return res.Destination, nil
}

type SignArgs struct {
	User string
	Groups []string
	Issuer string
}

func SignResponse(signingKey, samlResponseTmplBs []byte, args SignArgs) (signedResponse []byte, e error) {
	tmpl, e := template.New("").Parse(string(samlResponseTmplBs))
	if e != nil {
		return nil, e
	}

	samlResponseBuf := bytes.Buffer{}
	e = tmpl.Execute(&samlResponseBuf, args)
	if e != nil {
		return nil, e
	}
	samlResponseBs := samlResponseBuf.Bytes()

	// create some temp files for working with xmlsec1
	keyFile, e := ioutil.TempFile("", "")
	if e != nil {
		return
	}

	xmlSecInputFile, e := ioutil.TempFile("", "")
	if e != nil {
		return
	}

	xmlSecOutputFile, e := ioutil.TempFile("", "")
	if e != nil {
		return
	}

	defer removeFile(xmlSecInputFile.Name())
	defer removeFile(xmlSecOutputFile.Name())
	defer removeFile(keyFile.Name())

	// put the right content in the files
	_, e = keyFile.Write(signingKey)
	if e != nil {
		return
	}

	samlResponse := saml.Response{}
	e = xml.Unmarshal(samlResponseBs, &samlResponse)
	if e != nil {
		return
	}

	// IDs should be unique
	buf := make([]byte, 16)
	_, e = rand.Read(buf)
	if e != nil {
		return
	}
	samlID := hex.EncodeToString(buf)
	samlResponse.ID = samlID

	// Issuer should be set
	samlResponse.Issuer.Value = args.Issuer
	samlResponse.Assertion.Issuer.Value = args.Issuer

	// should be within auth window
	samlResponse.IssueInstant = time.Now().UTC()
	samlResponse.Assertion.IssueInstant = time.Now().UTC()
	samlResponse.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.NotOnOrAfter = time.Now().UTC().Add(5 * time.Minute)
	samlResponse.Assertion.Conditions.NotOnOrAfter = time.Now().UTC().Add(5 * time.Minute)

	// should be idp initiated (not in response to anything)
	samlResponse.InResponseTo = ""
	samlResponse.Assertion.Subject.SubjectConfirmation.SubjectConfirmationData.InResponseTo = ""

	e = xml.NewEncoder(xmlSecInputFile).Encode(&samlResponse)
	if e != nil {
		return
	}

	xmlSecInputFile.Close()

	// sign the assertion
	xmlSecPath, e := exec.LookPath("xmlsec1")
	if e != nil {
		return
	}

	command := exec.Command(
		xmlSecPath,
		"--sign",
		// the insecure flag is disabilng cert chain validation before signing. since
		// we know we trust this key and aren't relying on the cert chain for any security
		// properties, this isn't insecure, as the name might imply
		"--insecure",
		"--id-attr:ID",
		"Assertion",
		"--privkey-pem",
		keyFile.Name(),
		"--output",
		xmlSecOutputFile.Name(),
		xmlSecInputFile.Name(),
	)
	_, e = command.CombinedOutput()
	if e != nil {
		return
	}

	signedResponse, e = ioutil.ReadAll(xmlSecOutputFile)
	return
}

func removeFile(name string) {
	e := os.Remove(name)
	if e != nil {
		logrus.WithField("file", name).WithError(e).Error("failed to remove file. REMOVE MANUALLY!")
	}
}
