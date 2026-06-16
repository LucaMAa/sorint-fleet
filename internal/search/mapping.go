package search

const VehiclesMapping = `{
  "mappings": {
    "properties": {
      "id":            { "type": "keyword" },
      "license_plate": { "type": "keyword" },
      "brand":         { "type": "keyword" },
      "model":         { "type": "keyword" },
      "year":          { "type": "integer" },
      "color":         { "type": "keyword" },
      "fuel_type":     { "type": "keyword" },
      "status":        { "type": "keyword" },
      "assigned_to_id":{ "type": "keyword" },
      "assigned_at":   { "type": "date" },
      "mileage":       { "type": "integer" },
      "notes":         { "type": "text", "analyzer": "standard" },
      "jolly":         { "type": "boolean" },
      "jolly_duration":{ "type": "integer" },
      "created_at":    { "type": "date" },
      "updated_at":    { "type": "date" }
    }
  }
}`

const RequestsMapping = `{
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "user_id": { "type": "keyword" },
      "query_text": { "type": "text" },
      "created_at": { "type": "date" }
    }
  }
}`
