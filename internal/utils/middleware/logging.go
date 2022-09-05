package middleware

import (
	"bytes"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/willf/pad"

	"activecounter/pkg/log"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging is a middleware function that logs the each request.
func Logging() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now().UTC()
		path := ctx.Request.URL.Path

		// Read the Body content
		var bodyBytes []byte
		if ctx.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(ctx.Request.Body)
		}

		// Restore the io.ReadCloser to its original state
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// The basic informations.
		method := ctx.Request.Method
		ip := ctx.ClientIP()

		// log.Debugf("New request come in, path: %s, Method: %s, body `%s`", path, method, string(bodyBytes))
		blw := &bodyLogWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: ctx.Writer,
		}
		ctx.Writer = blw

		// Continue.
		ctx.Next()

		// Calculates the latency.
		end := time.Now().UTC()
		latency := end.Sub(start)
		code := ctx.Writer.Status()

		log.Infof("[GIN] statusCode:[%d] | %-13s | %-12s | %s %s", code, latency, ip,
			pad.Right(method, 5, ""), path)
	}
}
