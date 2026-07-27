package ginmiddleware

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

func getRequestValidationInput(
	req *http.Request,
	router routers.Router,
	options *Options,
) (*openapi3filter.RequestValidationInput, error) {
	route, pathParams, err := router.FindRoute(req)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return nil, fmt.Errorf("error validating route: %w", e)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return nil, fmt.Errorf("error validating route: %s", err.Error())
		}
	}

	reqValidationInput := openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	if options != nil {
		reqValidationInput.Options = &options.Options
		reqValidationInput.ParamDecoder = options.ParamDecoder
	}

	return &reqValidationInput, nil
}
