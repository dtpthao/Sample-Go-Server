package log

import (
	"time"

	"github.com/gin-gonic/gin"
)

func AccessLog(ctx *gin.Context) {
	start := time.Now()
	logger := Get()
	logger.Info().
		Str("IP", ctx.ClientIP()).
		Str("Method", ctx.Request.Method).
		Str("Path", ctx.Request.URL.Path).
		Str("Protocol", ctx.Request.Proto).
		Str("User-Agent", ctx.Request.UserAgent()).
		Str("Content-Type", ctx.Request.Header.Get("Content-Type")).
		Msg("REQUEST")
	ctx.Next()
	logger.Info().
		Str("IP", ctx.ClientIP()).
		Any("Status", ctx.Writer.Status()).
		Str("Path", ctx.Request.URL.Path).
		Dur("Latency", time.Since(start)).
		Str("Content-type", ctx.Writer.Header().Get("Content-Type")).
		Msg("RESPONSE")
}
