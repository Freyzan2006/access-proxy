package proxy


import (
    "net/http/httputil"
)

type Server interface {
	
}

type server struct {
	proxy *httputil.ReverseProxy
}

func NewServer(proxy *httputil.ReverseProxy) Server {
	return &server{
		proxy: proxy,
	}
}

