// Copyright © 2016 Apigee Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

type logRequest struct {
  Org string       `json:"org"`
  Env string       `json:"env"`
  Dep string       `json:"dep"`
  Namespace string `json:"namespace"`
  Tail int         `json:"tail"`
  TotalHits int    `json:"total_hits"`
}

//Error should be rendered when an error occurs
type Error struct {
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}

// ElasticSearchResponse is the response object of an elastic serach query
type ElasticSearchResponse struct {
  Took int `json:"took"`
  TimedOut bool `json:"timed_out"`
  Shards shards `json:"_shards"`
  Hits hits `json:"hits"`
}

type shards struct {
  Total int `json:"total"`
  Successful int `json:"successful"`
  Failed int `json:"failed"`
}

type hits struct {
  Total int `json:"total"`
  MaxScore float64 `json:"max_score"`
  Hits []hit `json:"hits"`
}

type hit struct {
  Index string `json:"_index"`
  Type string `json:"_type"`
  ID string `json:"_id"`
  Score float64 `json:"_score"`
  Source source `json:"_source"`
}

type source struct {
  Log string `json:"log"`
  Stream string `json:"stream"`
  K8sID string `json:"k8s_id"`
  Tag string `json:"tag"`
  Timestamp string `json:"@timestamp"`
}