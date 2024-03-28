package http_server

type IContext interface {
	GetRevealRequestsStructure()
}

type IHttpCallbacksListener interface {
	OnValidCaptcha(requestId string)
}
