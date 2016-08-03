# logpull

logpull is a basic Elasticsearch querying service to be run in Apigee's Kubernetes environment. It retrieves container logs.

## Environment
There are a few environment variables that are configurable, some are necessary.

`ELASTIC_SEARCH_HOST`: This is a required variable necessary to make queries against Elasticsearch.

`ELASTIC_SEARCH_PORT`: This isn't required, and can be omitted if you have a specific domain name for the client.

`PORT`: The port that the logpull service will listen on (default: `8000`)

`HIT_LIMIT`: The limit on how many Elasticsearch hits can be pulled by a single query (default: `1024`)

## Usage
With an Apigee org in hand and an environment and deployment made via Shipyard, retrieve deployment logs like so.

```sh
curl -H "Authorization: Bearer ${APIGEE_TOKEN}" "https://shipyard.apigee.com/logs/environments/${APIGEE_ORG}-${APIGEE_ENV}/deployments/<DEP_NAME>"
```

This will stream all of the available logs for the deployment since its creation. It doesn't diferentiate between an individual replica's logs, but merely dumps all discovered output in indexed order.

If you only want to see the 10 most recent log lines, for example, use the `tail` query string param.
