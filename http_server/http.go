package http_server

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/melardev/discord-message-protect/captchas"
	"github.com/melardev/discord-message-protect/core"
	"github.com/melardev/discord-message-protect/logging"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type MuxHttpServer struct {
	DefaultLogger    logging.ILogger
	AccessLogger     logging.ILogger
	CaptchaValidator captchas.ICaptchaValidator
	Context          IContext
	Listener         IHttpCallbacksListener
	webRouter        *mux.Router
	Config           *core.HttpConfig
	TemplateResponse string
}

type BlockedClient struct {
	ClientIp  string
	BlockedAt time.Time
}

func NewMuxHttpServer(config *core.Config, listener IHttpCallbacksListener) *MuxHttpServer {
	m := &MuxHttpServer{
		Config:       config.HttpConfig,
		Listener:     listener,
		AccessLogger: logging.NewFileLogger(filepath.Join(config.LogPath, "access.log")),
		DefaultLogger: logging.NewCompositeLogger(
			&logging.ConsoleLogger{},
			logging.NewFileLogger(filepath.Join(config.LogPath, "http.log")),
		),

		TemplateResponse: fmt.Sprintf(`<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Challenge</title>
    <CAPTCHA_JS>
</head>
<body>
<form method="post">
    Please solve the challenge
    <CAPTCHA_CHALLENGE>
    <br>
    <input type="submit" id="submit" value="Submit">
</form>
</body>
</html>`),
	}

	if config.HttpConfig.CaptchaService != "" {
		if config.HttpConfig.CaptchaService == "google" {
			m.CaptchaValidator = captchas.NewReCaptchaValidator(config.HttpConfig)
		} else {
			panic(fmt.Sprintf("%s captcha service is not supported, please use Google", config.HttpConfig.CaptchaService))
		}
	}

	return m

}

func (m *MuxHttpServer) HandleCaptcha(w http.ResponseWriter, req *http.Request) {
	if m.IsBlockedClient(req.RemoteAddr) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("An error occurred, that's all we know"))
		return
	}

	if req.Method == "GET" {
		m.GetCaptcha(w, req)
	} else if req.Method == "POST" {
		m.PostCaptcha(w, req)
	}
}

func (m *MuxHttpServer) GetCaptcha(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("An error occurred, that's all we know"))
		return
	}

	reqId := req.Form.Get("req_id")
	if reqId == "" {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("An error occurred, that's all we know"))
		return
	}

	// TODO: Implement me, static files or just write the html response here?
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(m.CaptchaValidator.AddHtml(m.TemplateResponse)))

}

func (m *MuxHttpServer) PostCaptcha(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		_, _ = w.Write([]byte("Error"))
		return
	}

	reCaptcha := req.Form.Get("g-recaptcha-response")

	success, _ := m.CaptchaValidator.ValidateCaptcha(reCaptcha)
	if !success {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("An error occurred, that's all we know"))
		return
	}

	reqId := req.Form.Get("req_id")
	if reqId != "" {
		m.Listener.OnValidCaptcha(reqId)
	}
}

func (m *MuxHttpServer) IsBlockedClient(clientIp string) bool {
	// TODO: Implement me
	return false
}

func (m *MuxHttpServer) Run() {
	m.webRouter = mux.NewRouter()
	m.webRouter.MatcherFunc(func(request *http.Request, match *mux.RouteMatch) bool {
		remoteIp := request.RemoteAddr
		if remoteIp == "127.0.0.1" || remoteIp == "localhost" {
			m.AccessLogger.Debug(fmt.Sprintf("Accessed by Host %s\n", remoteIp))
		} else {
			m.AccessLogger.Debug(fmt.Sprintf("Accessed by Ip: %s %s\n", remoteIp, request.RequestURI))
		}
		return false
	})

	challengePath := m.Config.ChallengePath
	if !strings.HasPrefix(challengePath, "/") {
		challengePath = "/" + challengePath
	}
	m.webRouter.HandleFunc(challengePath, m.HandleCaptcha).
		Methods("GET", "POST")

	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", m.Config.Port), m.webRouter)
}
