package ginmiddleware

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// GetGinContext gets the echo context from within requests. It returns
// nil if not found or wrong type.
func GetGinContext(c context.Context) *gin.Context {
	iface := c.Value(GinContextKey)
	if iface == nil {
		return nil
	}
	ginCtx, ok := iface.(*gin.Context)
	if !ok {
		return nil
	}
	return ginCtx
}

func GetUserData(c context.Context) interface{} {
	return c.Value(UserDataKey)
}

// attempt to get the MultiErrorHandler from the options. If it is not set,
// return a default handler
func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil {
		return defaultMultiErrorHandler
	}

	if options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

// defaultMultiErrorHandler returns a list of all of the errors.
// This method is called if there are no other methods defined on the options.
func defaultMultiErrorHandler(me openapi3.MultiError) error {
	return fmt.Errorf("multiple errors encountered: %s", me)
}
