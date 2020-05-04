package proxy

import (
	"bytes"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// Single page config.
type Config struct {
	PathPrefix     string
	Target         *url.URL
	ModifyResponse func(*http.Response) error
}

func (c Config) Handle(router *mux.Router) {
	router.Path(c.PathPrefix + "/").HandlerFunc(c.ServeHTTP)
	router.Path(c.PathPrefix + "/{w:.*}").Handler(c)
}

func (c Config) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Update the headers to allow for SSL redirection
	//request.URL.Host = c.Target.Host
	//request.URL.Scheme = c.Target.Scheme
	//request.Header.Set("X-Forwarded-Host", request.Header.Get("Host"))
	//request.Host = c.Target.Host

	proxy := httputil.NewSingleHostReverseProxy(c.Target)

	proxy.Director = c.director(proxy.Director)
	proxy.ModifyResponse = c.ModifyResponse

	proxy.ServeHTTP(response, request)
}

func (c Config) director(director func(req *http.Request)) func(req *http.Request) {
	return func(request *http.Request) {
		director(request)
		//Remove router path prefix.
		//prefix, err := mux.CurrentRoute(request).GetPathTemplate()
		//if err != nil {
		//	panic(err)
		//}
		request.URL.Path = request.URL.Path[len(c.PathPrefix):]
		request.Host = request.URL.Host

		//request.Header.Set("X-Forwarded-Host", c.Target.Host)
		//request.Header.Set("Origin", c.Target.Scheme+c.Target.Host) // Check
		//request.Header.Del("User-Agent") // Google on Chrome fix.
	}
}

func RewriteBody(response *http.Response, f func(body *[]byte) error) (err error) {
	b, _ := ioutil.ReadAll(response.Body)
	archived := isArchived(b)
	if archived {
		b, err = DecodeGZIP(b)
		if err != nil {
			return err
		}
	}

	err = f(&b)
	if err != nil {
		return err
	}

	if archived {
		b, err = EncodeGzip(b)
		if err != nil {
			return err
		}
	}
	response.Body = ioutil.NopCloser(bytes.NewReader(b))
	response.Header["Content-Length"] = []string{strconv.Itoa(len(b))}
	return
}

// Deprecate
func Redirect(response *http.Response) (err error) {
	code, err := strconv.Atoi(response.Status[:3])
	if err != nil {
		panic(err)
	}
	if code >= 300 && code < 400 {
		location, err := url.Parse(response.Header.Get("Location"))
		if err != nil {
			panic(err)
		}

		x, err := response.Location()
		if err != nil {
			return err
		}

		response.Header.Set("Location", x.Host+location.Path)
		return nil
	}
	return
}
