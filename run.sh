# Add connector
curl -i -X POST -H "Accept:application/json" -H  "Content-Type:application/json" http://localhost:8083/connectors/ -d @configs/postgres-source.json

# Check connector added
curl -s localhost:8083/connector-plugins | jq '.[].class'|egrep 'PostgresConnector|ElasticsearchSinkConnector'

# Check connector running
curl -s "http://localhost:8083/connectors?expand=info&expand=status" | \
  jq '. | to_entries[] | [ .value.info.type, .key, .value.status.connector.state,.value.status.tasks[].state,.value.info.config."connector.class"]|join(":|:")' | \
  column -s : -t| sed 's/\"//g'| sort