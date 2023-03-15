package server

import (
	"context"
	"embed"
	"net/http"
	"net/http/pprof"
	"strconv"
	"time"

	"techunicorn.com/udc-core/gettingStarted/pkg/app/server/common"
	"techunicorn.com/udc-core/gettingStarted/pkg/app/server/contracts"

	"github.com/betalixt/gorr"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

//go:embed static/*
var staticFiles embed.FS

func (a *app) startHTTP(portStr string) {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	router.SetTrustedProxies(nil)

	fileServer := http.FileServer(http.FS(staticFiles))

	// Index
	indexGroup := router.Group("/")
	indexGroup.GET("", func(ctx *gin.Context) {
		defer func(old string) {
			ctx.Request.URL.Path = old
		}(ctx.Request.URL.Path)

		ctx.Request.URL.Path = "/static/index" + ctx.Request.URL.Path
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
	indexGroup.GET("/index.css", func(ctx *gin.Context) {
		defer func(old string) {
			ctx.Request.URL.Path = old
		}(ctx.Request.URL.Path)

		ctx.Request.URL.Path = "/static/index" + ctx.Request.URL.Path
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
	indexGroup.GET("/favicon.ico", func(ctx *gin.Context) {
		defer func(old string) {
			ctx.Request.URL.Path = old
		}(ctx.Request.URL.Path)

		ctx.Request.URL.Path = "/static/index" + ctx.Request.URL.Path
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	// Swagger setup
	swaggerGroup := router.Group("/swagger")
	swaggerGroup.Any("/*all", func(ctx *gin.Context) {
		defer func(old string) {
			ctx.Request.URL.Path = old
		}(ctx.Request.URL.Path)

		ctx.Request.URL.Path = "/static" + ctx.Request.URL.Path
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})

	// Prometheus
	metricsGroup := router.Group("/metrics")
	metricsGroup.GET("", func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})

	// PProf setup
	pprofGroup := router.Group("pprof")
	pprofGroup.GET("/", gin.WrapF(pprof.Index))
	pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
	pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
	pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
	pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
	pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
	pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))

	// Health
	healthGroup := router.Group("/health")
	healthGroup.GET("", func(ctx *gin.Context) {
		ctx.Status(200)
	})
	healthGroup.GET("/complete", func(ctx *gin.Context) {
		err := a.impl.StatusCheck(context.Background())
		if err != nil {
			ctx.Status(500)
		} else {
			ctx.Status(200)
		}
	})

	// Application setup
	application := router.Group("")
	application.Use(a.traceRequestHandler)

	port, err := strconv.Atoi(portStr)
	if err != nil {
		a.lgr.Warn(
			"unable to parse provided port, setting port to default",
			zap.String("portConfig", portStr),
		)
		port = common.HTTPDefaultPort
	}
	if port < 0 {
		a.lgr.Warn(
			"port was configured was invalid, setting port to default",
		)
		port = common.HTTPDefaultPort
	}

	srv := http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	a.registerHTTPHandlers(application)
	a.registerCloser(func() {
		if err := srv.Close(); err != nil {
			a.lgr.Error("failed while closing http server", zap.Error(err))
		}
	})

	a.lgr.Info("http listening", zap.Int("port", port))

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.lgr.Error(
			"failed running router",
			zap.Error(err),
		)
	}
}

func (a *app) traceRequestHandler(ctx *gin.Context) {
	start := time.Now()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path
	query := ctx.Request.URL.RawQuery
	agent := ctx.Request.UserAgent()
	ip := ctx.ClientIP()

	traceparent := ctx.GetHeader("traceparent")
	c := a.ctxf.Create(traceparent)
	ctx.Set(contracts.InternalContextKey, c)
	ctx.Next()

	er := ctx.Errors.Last()
	if er != nil {
		if err, ok := er.Err.(*gorr.Error); ok {
			ctx.JSON(err.StatusCode, err)
		} else {
			ctx.JSON(500, gorr.NewUnexpectedError(er))
		}
	}

	status := ctx.Writer.Status()
	size := ctx.Writer.Size()
	end := time.Now()

	a.traceRequest(
		c,
		method,
		path,
		query,
		agent,
		ip,
		status,
		size,
		start,
		end,
		common.HTTPLable,
	)
}
