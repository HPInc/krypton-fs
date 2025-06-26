package rest

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	// REST request headers and expected header values.
	headerContentType         = "Content-Type"
	headerRequestID           = "request_id"
	contentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
	contentTypeJson           = "application/json"

	// Request parameters
	paramTenantID = "tenant_id"
	paramDeviceID = "device_id"
	paramFileID   = "id"
	paramMethod   = "method"
)

// getPathVariable gets & validates existence of string parameter
func getPathVariable(r *http.Request, variableName string,
	isRequired bool) (value string, err error) {
	vars := mux.Vars(r)
	value, ok := vars[variableName]
	if !ok {
		if isRequired {
			return "", ErrPathVariableMissing
		}
	}

	return value, nil
}

// getRequestPayload reads contents of the payload from the request body
func getRequestPayload(r *http.Request) (body []byte, err error) {
	body, err = io.ReadAll(r.Body)
	if err != nil {
		fsLogger.Error("Failed to read the request body!",
			zap.Error(err),
		)
		return nil, ErrPayloadRead
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return
}
