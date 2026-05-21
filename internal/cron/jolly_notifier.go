package cron

import (
	"log"
	"time"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/mailer"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/ws"

	"gorm.io/gorm"
)

type JollyNotifier struct {
	db             *gorm.DB
	assignmentRepo repository.VehicleAssignmentRepository
	vehicleRepo    repository.VehicleRepository
	interval       time.Duration
	stop           chan struct{}
}

func NewJollyNotifier() *JollyNotifier {
	return &JollyNotifier{
		db:             config.DB,
		assignmentRepo: repository.NewVehicleAssignmentRepository(),
		vehicleRepo:    repository.NewVehicleRepository(),
		interval:       1 * time.Hour,
		stop:           make(chan struct{}),
	}
}

func (j *JollyNotifier) Start() {
	log.Println("[JollyNotifier] started")
	go func() {
		j.run()
		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				j.run()
			case <-j.stop:
				log.Println("[JollyNotifier] stopped")
				return
			}
		}
	}()
}

func (j *JollyNotifier) Stop() {
	close(j.stop)
}

func (j *JollyNotifier) run() {
	log.Println("[JollyNotifier] checking expired jolly assignments...")

	var expired []model.VehicleAssignment

	err := j.db.
		Preload("Vehicle").
		Preload("User").
		Joins("JOIN vehicles ON vehicles.id = vehicle_assignments.vehicle_id").
		Where(`
			vehicle_assignments.ended_at IS NULL
			AND vehicles.jolly = true
			AND vehicle_assignments.started_at + (vehicles.jolly_duration || ' days')::interval <= ?
			AND vehicle_assignments.notified_at IS NULL
		`, time.Now()).
		Find(&expired).Error

	if err != nil {
		log.Printf("[JollyNotifier] query error: %v", err)
		return
	}

	log.Printf("[JollyNotifier] found %d expired jolly assignments", len(expired))

	for _, a := range expired {
		j.notify(a)

		now := time.Now()
		if err := j.db.Model(&model.VehicleAssignment{}).
			Where("id = ?", a.ID).
			Update("notified_at", now).Error; err != nil {
			log.Printf("[JollyNotifier] failed to mark notified assignment %s: %v", a.ID, err)
		}
	}
}

func (j *JollyNotifier) notify(a model.VehicleAssignment) {
	ws.Global.Broadcast("jolly_expired", map[string]interface{}{
		"assignment_id": a.ID,
		"vehicle":       a.Vehicle.Brand + " " + a.Vehicle.Model,
		"license_plate": a.Vehicle.LicensePlate,
		"user":          a.User.FirstName + " " + a.User.LastName,
		"email":         a.User.Email,
		"started_at":    a.StartedAt,
		"jolly_days":    a.Vehicle.JollyDuration,
	})

	go func() {
		if err := mailer.SendJollyExpiredEmail(a); err != nil {
			log.Printf("[JollyNotifier] email error: %v", err)
		}
	}()
}
