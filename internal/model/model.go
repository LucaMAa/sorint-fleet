package model

type Model struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"not null"`
	BrandID uint   `gorm:"index"`
	Brand   Brand  `gorm:"constraint:OnDelete:CASCADE"`
}
