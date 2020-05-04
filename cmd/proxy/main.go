package main

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/cuckmod/proxy/internal/proxy"
	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	donationalertsConfig().Handle(router)
	proxyConfig().Handle(router)
	if err := http.ListenAndServe(":6969", router); err != nil {
		panic(err)
	}
}

func proxyConfig() *proxy.Config {
	config := new(proxy.Config)
	config.PathPrefix = "/test"
	config.Target, _ = url.Parse("https://www.google.com")
	config.ModifyResponse = func(response *http.Response) (err error) {
		delete(response.Header, "Content-Security-Policy")
		delete(response.Header, "X-Xss-Protection")

		return proxy.RewriteBody(response, func(body *[]byte) error {
			index := bytes.Index(*body, []byte("</body>"))
			if index >= 0 {
				buffer := bytes.Buffer{}
				buffer.Write((*body)[:index])
				buffer.WriteString(`<h1 style="font-size: 5em; text-align: center;">HELLO WORLD</h1></body>`)
				buffer.Write((*body)[index:])
				*body = buffer.Bytes()
			}
			return nil
		})
	}
	return config
}

/*
localhost:6969/donationalerts?group_id=2&token=5fhDB9cw7CkuBkkKq7QW
*/
func donationalertsConfig() *proxy.Config {
	config := new(proxy.Config)
	config.PathPrefix = "/donationalerts"
	config.Target, _ = url.Parse("https://www.donationalerts.com/widget/alerts")
	config.ModifyResponse = func(response *http.Response) (err error) {
		delete(response.Header, "Content-Security-Policy")
		delete(response.Header, "X-Xss-Protection")

		err = proxy.RewriteBody(response, func(body *[]byte) error {
			index := bytes.Index(*body, []byte("</body>"))
			if index >= 0 {
				buffer := bytes.Buffer{}
				buffer.Write((*body)[:index])
				buffer.WriteString(`<script>console.log('Hello world');</script>`)
				buffer.Write((*body)[index:])
				*body = buffer.Bytes()
			}
			return nil
		})
		if err != nil {
			return err
		}
		return
	}
	return config
}

func widget(writer http.ResponseWriter, request *http.Request) {
	widget := getWidget(mux.Vars(request)["token"])
	target, err := url.Parse(widget["link"].(string))
	if err != nil {
		panic(err)
	}

	proxy, err := url.Parse("https://widget.cuckmod.com")
	if err != nil {
		panic(err)
	}
	proxy.Path = widget["service"].(string)
	proxy.RawQuery = target.RawQuery

	http.Redirect(writer, request, proxy.String(), http.StatusMovedPermanently)
}

func getWidget(token string) map[string]interface{} {
	req, err := http.NewRequest("GET", "localhost:8888/widget", nil)
	if err != nil {
		panic(err)
	}

	query := req.URL.Query()
	query.Set("token", token)
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp == nil {
		panic("response is nil")
	}

	m := map[string]interface{}{}
	err = jsonapi.UnmarshalPayload(resp.Body, &m)
	if err != nil {
		panic(err)
	}
	return m
}
