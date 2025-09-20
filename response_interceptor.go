package ginmiddleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type responseInterceptor struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

var _ gin.ResponseWriter = &responseInterceptor{}

func newResponseInterceptor(w gin.ResponseWriter) *responseInterceptor {
	return &responseInterceptor{
		ResponseWriter: w,
		body:           bytes.NewBufferString(""),
	}
}

func (w *responseInterceptor) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
