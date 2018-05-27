package api

import (
	"net/http"

	muxtrace "github.com/DataDog/dd-trace-go/contrib/gorilla/mux"
	"github.com/json-iterator/go"
	"gitlab.com/Cacophony/dhelpers"
	"gitlab.com/Cacophony/dhelpers/apihelper"
	"gitlab.com/Cacophony/dhelpers/middleware"
)

// New creates a new mux Web Service for reporting information about the SqsProcessor
func New() http.Handler {
	mux := muxtrace.NewRouter(muxtrace.WithServiceName("SqsProcessor-API"))

	mux.HandleFunc("/stats", getStats)

	// gzip response if accepted
	mux.Use(middleware.GzipMiddleware)

	return mux
}

func getStats(w http.ResponseWriter, _ *http.Request) {
	// gather data
	var result apihelper.SqsProcessorStatus
	result.Service = apihelper.GenerateServiceInformation()
	result.Available = true

	// return result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := jsoniter.NewEncoder(w).Encode(result)
	dhelpers.LogError(err)
}
