package server

import (
  "os"
  "io"
  "fmt"
  "errors"
  "net/http"
  "encoding/json"

  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "github.com/30x/authsdk"
)

// Port the port the server is listening on
var Port string
// ElasticSearchHost is the host name of the elastic search pod
var ElasticSearchHost string
// ElasticSearchPort  is the port of the elastic search pod
var ElasticSearchPort string

// DefaultPort is the default port to listen
const DefaultPort = "8000"

//Server struct
type Server struct {
	Router http.Handler
}

//NewServer creates a new server
func NewServer() (server *Server, err error) {
  router := mux.NewRouter()

  router.Path("/logs/environments/{org}-{env}/deployments/{dep}").Methods("GET").HandlerFunc(getDeploymentLogs)

  loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, router)

  server = &Server{
    Router: loggedRouter,
  }

  if Port = os.Getenv("PORT"); Port == "" {
    Port = DefaultPort
  }

  if ElasticSearchHost = os.Getenv("ELASTIC_SEARCH_HOST"); ElasticSearchHost == "" {
    return nil, errors.New("No ELASTIC_SEARCH_HOST set! Cannot query for logs without this!")
  }

  if ElasticSearchPort = os.Getenv("ELASTIC_SEARCH_PORT"); ElasticSearchPort == "" {
    return nil, errors.New("No ELASTIC_SEARCH_PORT set! Cannot query for logs without this!")
  }

  return server, nil
}

//Start the server
func (server *Server) Start() error {
  fmt.Println("logpull listening on", Port)
  return http.ListenAndServe(":"+Port, server.Router)
}

func getDeploymentLogs(w http.ResponseWriter, r *http.Request) {
  pathVars := mux.Vars(r)

  fmt.Printf("Validating org admin for %s\n", pathVars["org"])
  if !validateAdmin(pathVars["org"], w, r) {
    return
  }

  logReq := &logRequest{}
  logReq.Org = pathVars["org"]
  logReq.Env = pathVars["env"]
  logReq.Dep = pathVars["dep"]
  logReq.Namespace = logReq.Org + "-" + logReq.Env

  target := fmt.Sprintf("http://%s:%s/", ElasticSearchHost, ElasticSearchPort)

  req, err := http.NewRequest("GET", target, nil)
  if err != nil {
    writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
    return
  }

  res, err := http.DefaultClient.Do(req)
  if err != nil {
    writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
    return
  }

  if res.StatusCode < 200 && res.StatusCode >= 300 {
    writeErrorResponse(res.StatusCode, res.Status, w)
    return
  }

  w.Header().Set("X-Accel-Buffering", "no") // turn off proxy buffering
  w.Header().Set("X-Content-Type-Options", "nosniff")
  w.WriteHeader(http.StatusOK)

  _, err = io.Copy(w, res.Body)
  if err != nil {
    writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
    return
  }

  return
}

//validateAdmin Validate the requestor is an admin in the namepace.  If returns false, the caller should halt and return.  True if the request should continue.  TODO make this cleaner
func validateAdmin(org string, w http.ResponseWriter, r *http.Request) bool {

	//validate this user has a token and is org admin
	token, err := authsdk.NewJWTTokenFromRequest(r)

	if err != nil {
		message := fmt.Sprintf("Unable to find oAuth token %s", err)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	isAdmin, err := token.IsOrgAdmin(org)

	if err != nil {
		message := fmt.Sprintf("Unable to get permission token %s", err)
		writeErrorResponse(http.StatusUnauthorized, message, w)
		return false
	}

	//if not an admin, give access denied
	if !isAdmin {
		writeErrorResponse(http.StatusForbidden, fmt.Sprintf("You do not have admin permisison for org %s", org), w)
		return false
	}

	return true
}

//write a non 200 error response
func writeErrorResponse(statusCode int, message string, w http.ResponseWriter) {

	w.WriteHeader(statusCode)

	errorObject := Error{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errorObject)
}
