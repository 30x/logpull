package server

import (
  "os"
  "io/ioutil"
  "fmt"
  "errors"
  "net/http"
  "encoding/json"
  "strconv"

  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "github.com/30x/authsdk"
)

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

  ConfigureLogPull()

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

  tailStr := r.URL.Query().Get("tail")
  if tailStr == "" {
    tailStr = "0"
  }

  tail, err := strconv.Atoi(tailStr)
  if err != nil {
    writeErrorResponse(http.StatusBadRequest, err.Error(), w)
    return
  }
  logReq.Tail = tail

  // probe elastic search for any log entries for this deployment
  total, err := probeForLogs(logReq)
  if err != nil {
    writeErrorResponse(http.StatusInternalServerError, err.Error(), w)
    return
  }

  w.Header().Set("X-Accel-Buffering", "no") // turn off proxy buffering
  w.Header().Set("X-Content-Type-Options", "nosniff")
  w.WriteHeader(http.StatusOK)

  // use flusher for chunked encoding
  flusher, ok := w.(http.Flusher)
  if !ok {
    writeErrorResponse(http.StatusInternalServerError, "expected http.ResponseWriter to an http.Flusher", w)
    return
  }

  flusher.Flush()

  if total == 0 {
    w.Write([]byte("There are no logs available\n"))
  } else {
    logReq.TotalHits = total

    err = pullAndWriteLogs(w, flusher, logReq)
    if err != nil {
      w.Write([]byte(err.Error()))
    }
  }
}

func pullAndWriteLogs(w http.ResponseWriter, flusher http.Flusher, logReq *logRequest) error {
  fmt.Printf("Pulling and Writing logs: %v\n", logReq)
  if logReq.Tail == 0 { // get all existing logs
    if logReq.TotalHits >= HitLimit {
      remaining := 0

      // pull logs in chunks until we've pulled & written them all
      for remaining < logReq.TotalHits {
        res, err := pullLogBlock(logReq, HitLimit, remaining)
        if err != nil {
          return err
        }

        writeLogBlock(res, w)
        flusher.Flush()

        remaining += HitLimit
      }
    } else {
      // total hits is less than our hit pull limit, so grab them all from the beginning
      res, err := pullLogBlock(logReq, 0, logReq.TotalHits)
      if err != nil {
        return err
      }

      writeLogBlock(res, w)
      flusher.Flush()
    }
  } else {
    // get the last logReq.Tail number of log lines
    from := logReq.TotalHits - logReq.Tail
    res, err := pullLogBlock(logReq, from, logReq.Tail)
    if err != nil {
      return err
    }

    writeLogBlock(res, w)
    flusher.Flush()
  }

  fmt.Println("Done pulling and writing logs")

  return nil
}

// writes each log line of the pulled log hits
func writeLogBlock(res *ElasticSearchResponse, w http.ResponseWriter) {
  for _, hitObj := range res.Hits.Hits {
    w.Write([]byte(hitObj.Source.Log))
  }
}

// pulls a block of logs from 'from' until 'from+size'
func pullLogBlock(logReq *logRequest, from int, size int) (*ElasticSearchResponse, error) {
  target := fmt.Sprintf("http://%s:%s/_all/fluentd/_search?q=k8s_id:/%s-*/&size=%d&from=%d", ElasticSearchHost, ElasticSearchPort, logReq.Dep, size, from)

  fmt.Printf("Retrieving logs from %d to %d\n", from, from+size)
  res, err := http.Get(target)
  if err != nil {
    return nil, err
  }

  if res.StatusCode < 200 && res.StatusCode >= 300 {
    return nil, errors.New(res.Status)
  }

  // read entire elastic search response
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }

  // marshal response into ElasticSearchResponse struct
  logRes := ElasticSearchResponse{}
  err = json.Unmarshal(body, &logRes)
  if err != nil {
    return nil, err
  }

  return &logRes, nil
}

// returns total number of hits for the log query
func probeForLogs(logReq *logRequest) (int, error) {
  // query for deployment logs, using size=0 makes it quicker, we just want total number of hits
  target := fmt.Sprintf("http://%s:%s/_all/fluentd/_search?q=k8s_id:/%s-*/&size=0", ElasticSearchHost, ElasticSearchPort, logReq.Dep)

  fmt.Printf("Probing for logs: %v\n", logReq)
  res, err := http.Get(target)
  if err != nil {
    return -1, err
  }

  if res.StatusCode < 200 && res.StatusCode >= 300 {
    return -1, errors.New(res.Status)
  }

  // read entire elastic search response
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return -1, err
  }

  // marshal response into ElasticSearchResponse struct
  results := ElasticSearchResponse{}
  err = json.Unmarshal(body, &results)
  if err != nil {
    return -1, err
  }

  // return total number of hits for the query
  return results.Hits.Total, nil
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
