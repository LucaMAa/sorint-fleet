package seed

import (
	"fmt"
	"log"

	"sorint-fleet/internal/model"
	"sorint-fleet/internal/search"

	"gorm.io/gorm"
)

func ResetElasticsearch() error {
	if _, err := search.NewClientFromEnv(); err != nil {
		return err
	}
	if err := search.DeleteIndex("vehicles"); err != nil {
		log.Printf("warning: delete vehicles index failed: %v", err)
	}
	if err := search.DeleteIndex("requests"); err != nil {
		log.Printf("warning: delete requests index failed: %v", err)
	}
	return search.InitIndices()
}

func PopulateElasticsearch(db *gorm.DB) error {
	if search.ES == nil {
		return fmt.Errorf("Elasticsearch client not initialized")
	}
	if err := search.InitIndices(); err != nil {
		return err
	}
	var vehicles []model.Vehicle
	if err := db.Find(&vehicles).Error; err != nil {
		return err
	}
	for _, v := range vehicles {
		if err := search.IndexVehicle(v); err != nil {
			log.Printf("warning: failed to index vehicle %s: %v", v.ID.String(), err)
		}
	}
	return nil
}
