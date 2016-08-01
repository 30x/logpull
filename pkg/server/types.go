package server

type logRequest struct {
  Org string       `json:"org"`
  Env string       `json:"env"`
  Dep string       `json:"dep"`
  Namespace string `json:"namespace"`
}

//Error should be rendered when an error occurs
type Error struct {
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}