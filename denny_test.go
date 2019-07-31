package denny

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


type header struct {
	Key   string
	Value string
}
func performRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}


func TestSimpleRequest(t *testing.T) {
	signature := ""
	server := NewServer()
	server.Use(func(c *Context) {
		signature += "A"
		c.Next()
		signature += "B"
	})
	server.Use(func(c *Context) {
		signature += "C"
	})
	server.GET("/", func(c *Context) {
		signature += "D"
		c.String(http.StatusOK, signature)
	})
	server.NoRoute(func(c *Context) {
		signature += " X "
	})
	server.NoMethod(func(c *Context) {
		signature += " XX "
	})
	// RUN
	w := performRequest(server, "GET", "/")

	out, e := ioutil.ReadAll(w.Result().Body)

	assert.Equal(t, nil, e)
	assert.Equal(t, "ACD", string(out))
	// TEST
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ACDB", signature)
}
