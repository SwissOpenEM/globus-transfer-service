// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

const (
	ScicatKeyAuthScopes = "ScicatKeyAuth.Scopes"
)

// TransferPostJobParams defines parameters for TransferPostJob.
type TransferPostJobParams struct {
	SourceFacility string `form:"sourceFacility" json:"sourceFacility"`
	SourcePath     string `form:"sourcePath" json:"sourcePath"`
	DestFacility   string `form:"destFacility" json:"destFacility"`
	ScicatPid      string `form:"scicatPid" json:"scicatPid"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// request a transfer job
	// (POST /transfer)
	TransferPostJob(c *gin.Context, params TransferPostJobParams)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// TransferPostJob operation middleware
func (siw *ServerInterfaceWrapper) TransferPostJob(c *gin.Context) {

	var err error

	c.Set(ScicatKeyAuthScopes, []string{})

	// Parameter object where we will unmarshal all parameters from the context
	var params TransferPostJobParams

	// ------------- Required query parameter "sourceFacility" -------------

	if paramValue := c.Query("sourceFacility"); paramValue != "" {

	} else {
		siw.ErrorHandler(c, fmt.Errorf("Query argument sourceFacility is required, but not found"), http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "sourceFacility", c.Request.URL.Query(), &params.SourceFacility)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter sourceFacility: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Required query parameter "sourcePath" -------------

	if paramValue := c.Query("sourcePath"); paramValue != "" {

	} else {
		siw.ErrorHandler(c, fmt.Errorf("Query argument sourcePath is required, but not found"), http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "sourcePath", c.Request.URL.Query(), &params.SourcePath)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter sourcePath: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Required query parameter "destFacility" -------------

	if paramValue := c.Query("destFacility"); paramValue != "" {

	} else {
		siw.ErrorHandler(c, fmt.Errorf("Query argument destFacility is required, but not found"), http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "destFacility", c.Request.URL.Query(), &params.DestFacility)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter destFacility: %w", err), http.StatusBadRequest)
		return
	}

	// ------------- Required query parameter "scicatPid" -------------

	if paramValue := c.Query("scicatPid"); paramValue != "" {

	} else {
		siw.ErrorHandler(c, fmt.Errorf("Query argument scicatPid is required, but not found"), http.StatusBadRequest)
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "scicatPid", c.Request.URL.Query(), &params.ScicatPid)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter scicatPid: %w", err), http.StatusBadRequest)
		return
	}

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.TransferPostJob(c, params)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.POST(options.BaseURL+"/transfer", wrapper.TransferPostJob)
}

type TransferPostJobRequestObject struct {
	Params TransferPostJobParams
}

type TransferPostJobResponseObject interface {
	VisitTransferPostJobResponse(w http.ResponseWriter) error
}

type TransferPostJob200JSONResponse struct {
	// JobId the job id of the transfer job that was created
	JobId *string `json:"jobId,omitempty"`
}

func (response TransferPostJob200JSONResponse) VisitTransferPostJobResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type TransferPostJob400JSONResponse struct {
	// Message gives further context for failure
	Message *string `json:"message,omitempty"`
}

func (response TransferPostJob400JSONResponse) VisitTransferPostJobResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type TransferPostJob403JSONResponse struct {
	// Message gives further context to the reason why the request was denied
	Message *string `json:"message,omitempty"`
}

func (response TransferPostJob403JSONResponse) VisitTransferPostJobResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(403)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// request a transfer job
	// (POST /transfer)
	TransferPostJob(ctx context.Context, request TransferPostJobRequestObject) (TransferPostJobResponseObject, error)
}

type StrictHandlerFunc = strictgin.StrictGinHandlerFunc
type StrictMiddlewareFunc = strictgin.StrictGinMiddlewareFunc

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
}

// TransferPostJob operation middleware
func (sh *strictHandler) TransferPostJob(ctx *gin.Context, params TransferPostJobParams) {
	var request TransferPostJobRequestObject

	request.Params = params

	handler := func(ctx *gin.Context, request interface{}) (interface{}, error) {
		return sh.ssi.TransferPostJob(ctx, request.(TransferPostJobRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "TransferPostJob")
	}

	response, err := handler(ctx, request)

	if err != nil {
		ctx.Error(err)
		ctx.Status(http.StatusInternalServerError)
	} else if validResponse, ok := response.(TransferPostJobResponseObject); ok {
		if err := validResponse.VisitTransferPostJobResponse(ctx.Writer); err != nil {
			ctx.Error(err)
		}
	} else if response != nil {
		ctx.Error(fmt.Errorf("unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7xVTW8jNwz9K4Que5l43aYn34IWW7h7qNH0tsiB1tAzcsfirEjZawT+7wU144/YaZGg",
	"xZ4SC8PH9x6fqGfnedNzpKjiZs9OyOcUdP/oW9pQOXr0waN+pv1D1tYOQnQz1xLWlFzlIm7Izeyrn1Hv",
	"Hhbzu8+0d5XTfW/n2Af7fTgcKhfiig2gJvEp9BrYkP4gUXhYzGHFCbQl+LXjZRb4M2GUFSV4pLQNniYw",
	"V8hCAgMjUP6LopQyzNpSVDsOHCfWPmhn/f8BzBq6ym0pycDih8l0MnWHynFPEfvgZu5+Mp3cu8r1qG1x",
	"4qOOKPajZ9FbMXMF7DreDbQSfc0kGmIDx1JY81IAGwxRFBAG3wYtZ4Vmg4xM0XvOUcFzXIUmJ6phF7S9",
	"/OaDgMFh9OSKglR8mNdu5o7KFyz6Gy+LnoQbUkriZl/GcX7NlPbnaQrn5OkT+tAFtXNTEhLVbqYpU+XE",
	"8oG3+ldjCSibEkCBAeucCNEUYuMOh+rfei9Q23f0LWaUQrBxAcfiz4lOM8TAc4zktfhW0vtGTjWJ/j9u",
	"1CUPZTzvsKQEfhHqdzoyhmsx/wV4VQypUVFIYUmXoTS4WzJP1kx6jjIsgh+nU/vjOSrFkn3s+268cx/X",
	"Ym2fL/j0yZKoYahe89Ly+BrNNS8h1EeKlzcFtEWFHQr4RKiv0iyr5SWoZO9JZJW7bg+imJRqwBfIdtV/",
	"+k+CNiSCDd1KasKWBFY5aUsJCvo3LftghaHLid4ogjekrY1pR1Fhl9j+Pd78cbVUkCWjyawzWcysCuib",
	"UorYnXaIhCZi1xkYRqCUOA0G3H8/A5RH4igcYdfuL3WUGdcUw6sjPp3wck1e3StuGVYWSlAzSfyg0OKW",
	"hg6haUv3YyvJvr1KAwxPT7JNGhm22IUaOm4aqu9CLMCFhuTNBtPezdwR7CpVlVNsbK+601vxVNgeH9ay",
	"cq+e1C9PdtfGwmszfz9uc4FEnd0B0/LywTtvCjt3tkbeAmLL4ERfziAn6rdAn8ah8hnQgt1QpIQd2Auf",
	"Nlec2Crc4enwdwAAAP//WmKwlG0IAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
