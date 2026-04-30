package seed

import (
	"sorint-fleet/internal/model"

	"gorm.io/gorm"
)

func seedModels(db *gorm.DB) error {
	var toyota model.Brand
	if err := db.Where("name = ?", "Toyota").First(&toyota).Error; err != nil {
		return err
	}

	models := []model.Model{
		{Name: "Corolla", BrandID: toyota.ID},
		{Name: "Yaris", BrandID: toyota.ID},
		{Name: "RAV4", BrandID: toyota.ID},
	}

	for _, m := range models {
		var existing model.Model
		err := db.Where("name = ? AND brand_id = ?", m.Name, m.BrandID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&m).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
