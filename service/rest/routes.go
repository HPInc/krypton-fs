package rest

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Route - used to route REST requests received by the service.
type Route struct {
	Name        string           // Name of the route
	Method      string           // REST method
	Path        string           // Resource path
	HandlerFunc http.HandlerFunc // Request handler function.
}

type routes []Route

// List of Routes and corresponding handler functions registered
// with the router.
var registeredRoutes = routes{
	// Health method.
	Route{
		Name:        "GetHealth",
		Method:      http.MethodGet,
		Path:        "/health",
		HandlerFunc: GetHealthHandler,
	},

	// Metrics method.
	Route{
		Name:        "GetMetrics",
		Method:      http.MethodGet,
		Path:        "/metrics",
		HandlerFunc: promhttp.Handler().(http.HandlerFunc),
	},

	///////////////////////////////////////////////////////////////////////////
	//                   External API routes (device facing)                 //
	///////////////////////////////////////////////////////////////////////////

	// Create a record for the specified file in the files database and
	// return a pre-signed URL that the client can use to upload the file
	// to storage.
	Route{
		Name:        "CreateFile",
		Method:      http.MethodPost,
		Path:        "/api/v1/files",
		HandlerFunc: CreateFileHandler,
	},

	// Get information about the file corresponding to the specified file ID.
	Route{
		Name:        "GetFile",
		Method:      http.MethodGet,
		Path:        "/api/v1/files/{id:[0-9]+}",
		HandlerFunc: GetFileHandler,
	},

	///////////////////////////////////////////////////////////////////////////
	//                   Internal API routes (service facing)                //
	///////////////////////////////////////////////////////////////////////////

	// Run the database scavenger routine.
	Route{
		Name:        "RunScavenger",
		Method:      http.MethodPost,
		Path:        "/api/internal/v1/scavenger",
		HandlerFunc: ScavengeRequestHandler,
	},

	// Returns information about files matching the requested filter. Scoped
	// to a single tenant and device.
	Route{
		Name:        "ListFiles",
		Method:      http.MethodGet,
		Path:        "/api/internal/v1/files",
		HandlerFunc: ListFilesHandler,
	},

	// Delete the specified file from the files database and garbage collect it
	// from storage.
	Route{
		Name:        "DeleteFile",
		Method:      http.MethodDelete,
		Path:        "/api/internal/v1/files/{id:[0-9]+}",
		HandlerFunc: DeleteFileHandler,
	},

	// Produces presigned URL for GET, PUT, HEAD methods on an existing file.
	Route{
		Name:        "GetSignedUrl",
		Method:      http.MethodGet,
		Path:        "/api/internal/v1/files/{id:[0-9]+}/signed_url",
		HandlerFunc: GetSignedUrlHandler,
	},
}
