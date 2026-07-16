package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	platformmw "github.com/cairon-henrique-60/goMicrosserviceTemplateAPI/pkg/platform/middleware"
)

type Proxy struct{ services map[string]*url.URL }

func NewProxy(raw map[string]string) *Proxy {
	services := map[string]*url.URL{}
	for key, value := range raw {
		parsed, err := url.Parse(value)
		if err != nil {
			panic(err)
		}
		services[key] = parsed
	}
	return &Proxy{services}
}
func (p *Proxy) Public(service string) http.Handler            { return p.handler(service, "") }
func (p *Proxy) Protected(service, prefix string) http.Handler { return p.handler(service, prefix) }
func (p *Proxy) handler(service, prefix string) http.Handler {
	target := p.services[service]
	proxy := httputil.NewSingleHostReverseProxy(target)
	original := proxy.Director
	proxy.Director = func(req *http.Request) {
		original(req)
		if prefix != "" && !strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = prefix + req.URL.Path
		}
		if user, ok := platformmw.UserFromContext(req.Context()); ok {
			req.Header.Set("X-User-ID", user.UserID)
			req.Header.Set("X-User-Email", user.Email)
			req.Header.Set("X-User-Roles", strings.Join(user.Roles, ","))
		}
	}
	return proxy
}
