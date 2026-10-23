# Gin Middleware

⚠️ This README may be for the latest development version, which may contain unreleased changes. Please ensure you're looking at the README for the latest release version.

Middleware for the [Gin web server](https://github.com/gin-gonic/gin) for use with [oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

## Validated request body

The middleware enables kin-openapi's `openapi3filter.Options.KeepValidatedValue` by default. After a request body passes validation, `c.Request.Body` stays readable (with the original bytes, unless kin-openapi rewrote it to apply schema defaults) and additionally implements `interface{ ValidatedValue() any }`, returning the decoded value that was validated. oapi-codegen strict servers that support it can bind that value instead of decoding the body a second time, so the handler sees exactly what was validated (for example, `encoding/json`'s case-insensitive member matching can no longer make a request validate as one value but bind as another).

To opt out, for instance to limit memory retention for very large request bodies, set `Options.DiscardValidatedValue = true`. This setting overrides `Options.Options.KeepValidatedValue`.

Any middleware placed after the validator that replaces `c.Request.Body` (decompression, size limiting, logging wrappers and the like) hides the validated value from handlers. Register such middleware before the validator.

Licensed under the Apache-2.0.
