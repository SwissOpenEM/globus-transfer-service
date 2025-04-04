package api

import (
	"embed"
	"net"
	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//go:embed openapi.yaml
var swaggerYAML embed.FS

func NewServer(api *ServerHandler, port uint, scicatUrl string) (*http.Server, error) {
	r := gin.New()

	r.GET("/openapi.yaml", func(c *gin.Context) {
		http.FileServer(http.FS(swaggerYAML)).ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.URL("/openapi.yaml")))

	r.Use(
		ScicatTokenAuthMiddleware(scicatUrl),
	)

	RegisterHandlers(r, NewStrictHandler(api, []StrictMiddlewareFunc{}))

	return &http.Server{
		Handler: r,
		Addr:    net.JoinHostPort("0.0.0.0", fmt.Sprint(port)),
	}, nil
}
