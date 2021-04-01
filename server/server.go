package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/willbtlr/imperson8/saml"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"sync"
)

const responseFormTmpl = `
<html>
	<head>
		<style>
			body {
				background-color: black;
				color: lightgreen;
				font-size: 64;
			}
		</style>
	</head>
	<body>
		<marquee scrolldelay=0 scrollamount=40>Hacking in progress...</marquee>	
		<form id="responseForm" method="POST" action="{{.Url}}">
			<input type="hidden" name="SAMLResponse" value="{{.Assertion}}"/>
		</form>
		<script type="text/javascript">
			setTimeout(() => {
			document.getElementById("responseForm").submit();
			}, 3000);
		</script>
	</body>
</html>`

type tmplArgs struct {
	Url       string
	Assertion string
}

type Server struct {
	Port     int
	Done     chan bool
	response []byte
	token    string
	lock     sync.RWMutex
}

func (s *Server) SetResponse(response []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.response = response
}

func (s *Server) SetToken() (string, error) {
	tokenBs := make([]byte, 16)
	_, e := rand.Read(tokenBs)
	if e != nil {
		return "", e
	}

	token := hex.EncodeToString(tokenBs)

	s.lock.Lock()
	s.token = token
	s.lock.Unlock()

	return token, nil
}

func (s *Server) Start() error {
	s.Done = make(chan bool)

	htmlTmpl, e := template.New("").Parse(responseFormTmpl)
	if e != nil {
		return e
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		s.lock.RLock()
		response := s.response
		token := s.token
		s.lock.RUnlock()

		// token to prevent local attacker from racing us
		providedToken := req.URL.Query().Get("token")
		if token != providedToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		destination, e := saml.Destination(s.response)
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		e = htmlTmpl.Execute(w, tmplArgs{
			Url:       destination,
			Assertion: base64.StdEncoding.EncodeToString(response),
		})
		if e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}

		s.Done <- true
	})

	go func() {
		e := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", s.Port), http.DefaultServeMux)
		if e != nil {
			logrus.WithError(e).Fatal("cant start server")
		}
	}()

	return nil
}
