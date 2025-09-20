package ginmiddleware

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
)

type GinContextKeyType string
type UserDataKeyType string

const (
	GinContextKey GinContextKeyType = "oapi-codegen/gin-context"
	UserDataKey   UserDataKeyType   = "oapi-codegen/user-data"
)

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(c *gin.Context, message string, statusCode int)

// MultiErrorHandler is called when oapi returns a MultiError type
type MultiErrorHandler func(openapi3.MultiError) error

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	ErrorHandler      ErrorHandler
	Options           openapi3filter.Options
	ParamDecoder      openapi3filter.ContentParameterDecoder
	UserData          interface{}
	MultiErrorHandler MultiErrorHandler
	// SilenceServersWarning allows silencing a warning for https://github.com/deepmap/oapi-codegen/issues/882 that reports when an OpenAPI spec has `spec.Servers != nil`
	SilenceServersWarning bool
}
