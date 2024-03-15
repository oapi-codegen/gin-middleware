package ginmiddleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func getRequestContext(
	c *gin.Context,
	options *Options,
) context.Context {
	requestContext := context.WithValue(context.Background(), GinContextKey, c)
	if options != nil {
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	return requestContext
}
