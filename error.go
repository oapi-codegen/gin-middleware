package ginmiddleware

import (
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/routers"
	"github.com/gin-gonic/gin"
)

func handleValidationError(
	c *gin.Context,
	err error,
	options *Options,
	generalStatusCode int,
) {
	var errorHandler ErrorHandler
	// if an error handler is provided, use that
	if options != nil && options.ErrorHandler != nil {
		errorHandler = options.ErrorHandler
	} else {
		errorHandler = func(c *gin.Context, message string, statusCode int) {
			c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
		}
	}

	if errors.Is(err, routers.ErrPathNotFound) {
		errorHandler(c, err.Error(), http.StatusNotFound)
	} else {
		errorHandler(c, err.Error(), generalStatusCode)
	}

	// in case the handler didn't internally call Abort, stop the chain
	c.Abort()
}
