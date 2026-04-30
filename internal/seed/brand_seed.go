package seed

import (
	"sorint-fleet/internal/model"

	"gorm.io/gorm"
)

func seedBrands(db *gorm.DB) error {
	brands := []model.Brand{
		{Name: "Toyota"},
		{Name: "BMW"},
		{Name: "Audi"},
		{Name: "Mercedes"},
		{Name: "Volkswagen"},
		{Name: "Fiat"},
		{Name: "Tesla"},
	}

	for _, b := range brands {
		var existing model.Brand
		err := db.Where("name = ?", b.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&b).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
