// Copyright 2021 DeepMap, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ginmiddleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

// OapiResponseValidatorFromYamlFile creates a validator middleware from a YAML file path
func OapiResponseValidatorFromYamlFile(path string) (gin.HandlerFunc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", path, err)
	}

	swagger, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s as Swagger YAML: %s",
			path, err)
	}
	return OapiRequestValidator(swagger), nil
}

// OapiRequestValidator is an gin middleware function which validates incoming HTTP requests
// to make sure that they conform to the given OAPI 3.0 specification. When
// OAPI validation fails on the request, we return an HTTP/400 with error message
func OapiResponseValidator(swagger *openapi3.T) gin.HandlerFunc {
	return OapiResponseValidatorWithOptions(swagger, nil)
}

// OapiResponseValidatorWithOptions creates a validator from a swagger object, with validation options
func OapiResponseValidatorWithOptions(swagger *openapi3.T, options *Options) gin.HandlerFunc {
	if swagger.Servers != nil && (options == nil || !options.SilenceServersWarning) {
		log.Println("WARN: OapiResponseValidatorWithOptions called with an OpenAPI spec that has `Servers` set. This may lead to an HTTP 400 with `no matching operation was found` when sending a valid request, as the validator performs `Host` header validation. If you're expecting `Host` header validation, you can silence this warning by setting `Options.SilenceServersWarning = true`. See https://github.com/deepmap/oapi-codegen/issues/882 for more information.")
	}

	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		err := ValidateResponseFromContext(c, router, options)
		if err != nil {
			if options != nil && options.ErrorHandler != nil {
				options.ErrorHandler(c, err.Error(), http.StatusInternalServerError)
				// in case the handler didn't internally call Abort, stop the chain
				c.Abort()
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}
	}
}

type responseInterceptor struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseInterceptor) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// ValidateResponseFromContext is called from the middleware above and actually does the work
// of validating a response.
func ValidateResponseFromContext(c *gin.Context, router routers.Router, options *Options) error {
	req := c.Request
	route, pathParams, err := router.FindRoute(req)

	// We failed to find a matching route for the request.
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return errors.New(e.Reason)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return fmt.Errorf("error validating route: %s", err.Error())
		}
	}

	reqValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Pass the gin context into the request validator, so that any callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), GinContextKey, c)

	if options != nil {
		reqValidationInput.Options = &options.Options
		reqValidationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	// wrap the response writer in a bodyWriter so we can capture the response body
	bw := &responseInterceptor{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
	c.Writer = bw

	// Call the next handler in the chain, which will actually process the request
	c.Next()

	// capture the response status and body
	status := c.Writer.Status()
	body := io.NopCloser(bytes.NewReader(bw.body.Bytes()))

	rspValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqValidationInput,
		Status:                 status,
		Header:                 c.Writer.Header(),
		Body:                   body,
	}

	if options != nil {
		rspValidationInput.Options = &options.Options
	}

	err = openapi3filter.ValidateResponse(requestContext, rspValidationInput)

	if err != nil {
		// restore the original response writer
		c.Writer = bw.ResponseWriter

		me := openapi3.MultiError{}
		if errors.As(err, &me) {
			errFunc := getMultiErrorHandlerFromOptions(options)
			return errFunc(me)
		}

		switch e := err.(type) {
		case *openapi3filter.ResponseError:
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(e.Error(), "\n")
			return fmt.Errorf("error in openapi3filter.ResponseError: %s", errorLines[0])
		case *openapi3filter.SecurityRequirementsError:
			return fmt.Errorf("error in openapi3filter.SecurityRequirementsError: %s", e.Error())
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return fmt.Errorf("error validating response: %w", err)
		}
	}

	// the response is valid, so write the captured response body to the original response writer
	_, err = bw.ResponseWriter.Write(bw.body.Bytes())
	if err != nil {
		return fmt.Errorf("error writing response body: %w", err)
	}

	return nil
}
