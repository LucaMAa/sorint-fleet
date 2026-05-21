package mailer

import (
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
)

func SendJollyExpiredEmail(a model.VehicleAssignment) error {
	body, err := render("jolly_expired.html", dto.JollyExpiredDto{
		FirstName:    a.User.FirstName,
		LastName:     a.User.LastName,
		Vehicle:      a.Vehicle.Brand + " " + a.Vehicle.Model,
		LicensePlate: a.Vehicle.LicensePlate,
		StartedAt:    a.StartedAt.Format("02/01/2006"),
		JollyDays:    a.Vehicle.JollyDuration,
	})
	if err != nil {
		return err
	}

	adminEmail := adminEmailFromEnv()
	return Send(adminEmail, "Auto jolly scaduta — "+a.Vehicle.LicensePlate, body)
}
