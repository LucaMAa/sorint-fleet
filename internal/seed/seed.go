package seed

import "gorm.io/gorm"

func Run(db *gorm.DB) error {
	if err := seedBrands(db); err != nil {
		return err
	}

	if err := seedModels(db); err != nil {
		return err
	}

	return nil
}
