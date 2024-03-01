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
	_ "embed"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_response_spec.yaml
var testResponseSchema []byte

func TestOapiResponseValidator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	swagger, err := openapi3.NewLoader().LoadFromData(testResponseSchema)
	require.NoError(t, err, "Error initializing swagger")

	// Create a new echo router
	g := gin.New()

	// Set up an authenticator to check authenticated function. It will allow
	// access to "someScope", but disallow others.
	options := Options{
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			c.String(statusCode, "test: "+message)
		},
		Options: openapi3filter.Options{
			AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
			IncludeResponseStatus: true,
		},
		UserData: "hi!",
	}

	// Install our OpenApi based request validator
	g.Use(OapiResponseValidatorWithOptions(swagger, &options))

	tests := []struct {
		name        string
		operationID string
	}{
		{
			name:        "GET /resource",
			operationID: "getResource",
		},
	}

	// getResource
	testGetResource := func(t *testing.T, g *gin.Engine) {
		var body string
		var statusCode int
		var contentType string

		// Install a request handler for /resource.
		g.GET("/resource", func(c *gin.Context) {
			c.Data(statusCode, contentType, []byte(body))
		})

		tests := []struct {
			name        string
			body        string
			status      int
			contentType string
			wantRsp     string
			wantStatus  int
		}{
			// Let's send a good response, it should pass
			{
				name:        "good response: good status: 200",
				body:        `{"name": "Wilhelm Scream", "id": 7}`,
				status:      http.StatusOK,
				contentType: "application/json",
				wantRsp:     `{"name":"Wilhelm Scream", "id":7}`,
				wantStatus:  http.StatusOK,
			},
			// And for 404, it should pass
			{
				name:        "good response: good status: 404",
				body:        `{"message": "couldn't find the resource"}`,
				status:      http.StatusNotFound,
				contentType: "application/json",
				wantRsp:     `{"message": "couldn't find the resource"}`,
				wantStatus:  http.StatusNotFound,
			},
			// And for 500, it should pass
			{
				name:        "good response: good status: 500",
				body:        `{"message": "internal server error"}`,
				status:      http.StatusInternalServerError,
				contentType: "application/json",
				wantRsp:     `{"message": "internal server error"}`,
				wantStatus:  http.StatusInternalServerError,
			},
			// Let's send a bad response, it should fail
			{
				name:        "bad response: good status",
				body:        `{"name": "Wilhelm Scream", "id": "not a number"}`,
				status:      http.StatusOK,
				contentType: "application/json",
				wantRsp:     `test: error in openapi3filter.ResponseError: response body doesn't match schema: Error at "/id": value must be an integer`,
				wantStatus:  http.StatusInternalServerError,
			},
			// And for 404, it should fail
			{
				name:        "bad response: missing required property: good status: 404",
				body:        `{}`,
				status:      http.StatusNotFound,
				contentType: "application/json",
				wantRsp:     `test: error in openapi3filter.ResponseError: response body doesn't match schema: Error at "/message": property "message" is missing`,
				wantStatus:  http.StatusInternalServerError,
			},
			// Let's send a good response, but with a bad status, it should fail
			{
				name:        "good response: bad status",
				body:        `{"name": "Wilhelm Scream", "id": 7}`,
				status:      http.StatusCreated,
				contentType: "application/json",
				wantRsp:     `test: error in openapi3filter.ResponseError: status is not supported`,
				wantStatus:  http.StatusInternalServerError,
			},
			// Let's send a good response, but with a bad content type, it should fail
			{
				name:        "good response: bad content type",
				body:        `{"name": "Wilhelm Scream", "id": 7}`,
				status:      http.StatusOK,
				contentType: "text/plain",
				wantRsp:     `test: error in openapi3filter.ResponseError: response header Content-Type has unexpected value: "text/plain"`,
				wantStatus:  http.StatusInternalServerError,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body = tt.body
				statusCode = tt.status
				contentType = tt.contentType

				rec := doGet(t, g, "http://deepmap.ai/resource")
				assert.Equal(t, tt.wantStatus, rec.Code)
				if tt.wantStatus == http.StatusOK {
					switch tt.contentType {
					case "application/json":
						assert.JSONEq(t, tt.wantRsp, rec.Body.String())
					default:
						assert.Equal(t, tt.wantRsp, rec.Body.String())
					}
				} else {
					assert.Equal(t, tt.wantRsp, rec.Body.String())
				}
			})
		}
	}

	// createResource
	testCreateResource := func(t *testing.T, g *gin.Engine) {
		var body string
		var statusCode int
		var contentType string

		// Install a request handler for /resource.
		g.POST("/resource", func(c *gin.Context) {
			c.Data(statusCode, contentType, []byte(body))
		})

		tests := []struct {
			name        string
			body        string
			status      int
			contentType string
			wantRsp     string
			wantStatus  int
		}{
			// Let's send a good response, it should pass
			{
				name:        "good response: good status: 201",
				body:        `{"name": "Wilhelm Scream", "id": 7}`,
				status:      http.StatusCreated,
				contentType: "application/json",
				wantRsp:     `{"name":"Wilhelm Scream", "id":7}`,
				wantStatus:  http.StatusCreated,
			},
			// Let's send a good response, but with a bad status, it should fail
			{
				name:        "good response: bad status: 200",
				body:        `{"name": "Wilhelm Scream", "id": 7}`,
				status:      http.StatusOK,
				contentType: "application/json",
				wantRsp:     `test: error in openapi3filter.ResponseError: status is not supported`,
				wantStatus:  http.StatusInternalServerError,
			},
			// Let's send a good response, with different content type, it should pass
			{
				name:        "good response: good status: 504",
				body:        "Gateway Timeout",
				status:      http.StatusGatewayTimeout,
				contentType: "text/plain",
				wantRsp:     "Gateway Timeout",
				wantStatus:  http.StatusGatewayTimeout,
			},
			// Let's send a good response, but with a bad content type, it should fail
			{
				name:        "good response: bad content type",
				body:        `{"message":"timed out waiting for upstream server to respond"}`,
				status:      http.StatusGatewayTimeout,
				contentType: "application/json",
				wantRsp:     `test: error in openapi3filter.ResponseError: response header Content-Type has unexpected value: "application/json"`,
				wantStatus:  http.StatusInternalServerError,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body = tt.body
				statusCode = tt.status
				contentType = tt.contentType

				rec := doPost(t, g, "http://deepmap.ai/resource", gin.H{"name": "Wilhelm Scream"})
				assert.Equal(t, tt.wantStatus, rec.Code)
				if tt.wantStatus == http.StatusOK {
					switch tt.contentType {
					case "application/json":
						assert.JSONEq(t, tt.wantRsp, rec.Body.String())
					default:
						assert.Equal(t, tt.wantRsp, rec.Body.String())
					}
				} else {
					assert.Equal(t, tt.wantRsp, rec.Body.String())
				}
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operationID {
			case "getResource":
				testGetResource(t, g)
			case "createResource":
				testCreateResource(t, g)
			}
		})
	}

}
