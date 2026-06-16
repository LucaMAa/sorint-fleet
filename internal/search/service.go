package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func InitIndices() error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	if err := CreateIndexIfNotExists("vehicles", VehiclesMapping); err != nil {
		return err
	}
	if err := CreateIndexIfNotExists("requests", RequestsMapping); err != nil {
		return err
	}
	return nil
}

func vehicleDocFromModel(v model.Vehicle) map[string]interface{} {
	doc := map[string]interface{}{
		"id":             v.ID.String(),
		"license_plate":  v.LicensePlate,
		"brand":          v.Brand,
		"model":          v.Model,
		"year":           v.Year,
		"color":          v.Color,
		"fuel_type":      v.FuelType,
		"status":         v.Status,
		"mileage":        v.Mileage,
		"notes":          v.Notes,
		"jolly":          v.Jolly,
		"jolly_duration": v.JollyDuration,
		"created_at":     v.CreatedAt,
		"updated_at":     v.UpdatedAt,
	}
	if v.AssignedToID != nil {
		doc["assigned_to_id"] = v.AssignedToID.String()
	}
	if v.AssignedAt != nil {
		doc["assigned_at"] = v.AssignedAt
	}
	return doc
}

func IndexVehicle(v model.Vehicle) error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	doc := vehicleDocFromModel(v)
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	req := esapi.IndexRequest{
		Index:      "vehicles",
		DocumentID: v.ID.String(),
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error indexing vehicle: %s", string(body))
	}
	return nil
}

func DeleteVehicle(id string) error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	req := esapi.DeleteRequest{
		Index:      "vehicles",
		DocumentID: id,
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error deleting vehicle: %s", string(body))
	}
	return nil
}

func IndexAllVehicles() error {
	var vehicles []model.Vehicle
	if err := config.DB.Find(&vehicles).Error; err != nil {
		return err
	}
	for _, v := range vehicles {
		if err := IndexVehicle(v); err != nil {
			log.Printf("indexing vehicle %s failed: %v", v.ID.String(), err)
		}
	}
	return nil
}

func requestDocFromModel(r model.Request) map[string]interface{} {
	doc := map[string]interface{}{
		"id":         r.ID.String(),
		"user_id":    r.UserID.String(),
		"query_text": r.QueryText,
		"created_at": r.CreatedAt,
	}
	return doc
}

func IndexRequest(r model.Request) error {
	if ES == nil {
		return fmt.Errorf("ES client not initialized")
	}
	doc := requestDocFromModel(r)
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	req := esapi.IndexRequest{
		Index:      "requests",
		DocumentID: r.ID.String(),
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}
	res, err := req.Do(context.Background(), ES)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error indexing request: %s", string(body))
	}
	return nil
}

func detectVehicleSearchIntent(q string) (fuelType string, wantsCarplay bool, wantsParking bool) {
	normalized := strings.ToLower(q)
	if strings.Contains(normalized, "elettr") || strings.Contains(normalized, "electric") || strings.Contains(normalized, "ev") {
		fuelType = "elettrico"
	}
	if strings.Contains(normalized, "benz") || strings.Contains(normalized, "gasol") {
		fuelType = "benzina"
	}
	if strings.Contains(normalized, "diesel") {
		fuelType = "diesel"
	}
	if strings.Contains(normalized, "ibrid") || strings.Contains(normalized, "hybrid") {
		fuelType = "ibrido"
	}
	wantsCarplay = strings.Contains(normalized, "carplay") || strings.Contains(normalized, "car play")
	wantsParking = strings.Contains(normalized, "parking") || strings.Contains(normalized, "parchegg") || strings.Contains(normalized, "park")
	return
}

func SearchRequests(queryText string, size int) ([]map[string]interface{}, error) {
	if ES == nil {
		return nil, fmt.Errorf("ES client not initialized")
	}
	q := strings.TrimSpace(queryText)
	var dsl map[string]interface{}
	if q == "" {
		dsl = map[string]interface{}{
			"size":  size,
			"query": map[string]interface{}{"match_all": map[string]interface{}{}},
			"sort":  []interface{}{map[string]interface{}{"created_at": map[string]interface{}{"order": "desc"}}},
		}
	} else {
		dsl = map[string]interface{}{
			"size": size,
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{
						map[string]interface{}{
							"multi_match": map[string]interface{}{
								"query":     q,
								"fields":    []string{"query_text"},
								"type":      "best_fields",
								"fuzziness": "AUTO",
							},
						},
					},
				},
			},
		}
	}
	bodyBytes, _ := json.Marshal(dsl)
	res, err := ES.Search(ES.Search.WithContext(context.Background()), ES.Search.WithBody(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", string(b))
	}
	var parsed map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	hits := []map[string]interface{}{}
	if h, ok := parsed["hits"].(map[string]interface{}); ok {
		if hs, ok := h["hits"].([]interface{}); ok {
			for _, item := range hs {
				if it, ok := item.(map[string]interface{}); ok {
					if src, ok := it["_source"].(map[string]interface{}); ok {
						hits = append(hits, src)
					}
				}
			}
		}
	}
	return hits, nil
}

func SearchVehicles(queryText string, desiredBody string, carplay *bool, parking *bool, size int) ([]map[string]interface{}, error) {
	if ES == nil {
		return nil, fmt.Errorf("ES client not initialized")
	}
	var dsl map[string]interface{}
	q := strings.TrimSpace(queryText)
	fuelType, inferredCarplay, inferredParking := detectVehicleSearchIntent(q)
	if q == "" {
		dsl = map[string]interface{}{
			"size":  size,
			"query": map[string]interface{}{"match_all": map[string]interface{}{}},
			"sort":  []interface{}{map[string]interface{}{"created_at": map[string]interface{}{"order": "desc"}}},
		}
	} else {
		must := []interface{}{}
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  q,
				"fields": []string{"brand^3", "model^3", "fuel_type^5", "notes^2"},
				"type":   "best_fields",
			},
		})

		if fuelType != "" {
			must = append(must, map[string]interface{}{
				"match": map[string]interface{}{"fuel_type": map[string]interface{}{"query": fuelType, "boost": 5}},
			})
		}
		if inferredCarplay || (carplay != nil && *carplay) {
			must = append(must, map[string]interface{}{
				"match": map[string]interface{}{"notes": map[string]interface{}{"query": "carplay", "boost": 4}},
			})
		}
		if inferredParking || (parking != nil && *parking) {
			must = append(must, map[string]interface{}{
				"match": map[string]interface{}{"notes": map[string]interface{}{"query": "parking", "boost": 4}},
			})
		}

		should := []interface{}{
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  q,
					"fields": []string{"brand", "model", "notes", "fuel_type"},
					"type":   "best_fields",
				},
			},
		}

		boolQuery := map[string]interface{}{
			"must":                 must,
			"should":               should,
			"minimum_should_match": 1,
		}

		dsl = map[string]interface{}{
			"size": size,
			"query": map[string]interface{}{
				"bool": boolQuery,
			},
		}
	}
	bodyBytes, _ := json.Marshal(dsl)
	res, err := ES.Search(ES.Search.WithContext(context.Background()), ES.Search.WithBody(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", string(b))
	}
	var parsed map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	hits := []map[string]interface{}{}
	if h, ok := parsed["hits"].(map[string]interface{}); ok {
		if hs, ok := h["hits"].([]interface{}); ok {
			for _, item := range hs {
				if it, ok := item.(map[string]interface{}); ok {
					if src, ok := it["_source"].(map[string]interface{}); ok {
						hits = append(hits, src)
					}
				}
			}
		}
	}
	return hits, nil
}
