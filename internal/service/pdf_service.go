package service

import (
	"errors"
	"fmt"
	"os"
	"sorint-fleet/internal/gotenberg"
	"sorint-fleet/internal/model"
)

type PDFService interface {
    AssignmentPDF(vehicle *model.Vehicle) ([]byte, error)
}

type pdfService struct {
    generator  *gotenberg.Generator
    signerName string
}

func NewPDFService(generator *gotenberg.Generator, signerName string) *pdfService {
	if signerName == "" {
		signerName = os.Getenv("PDF_SIGNER_NAME")
	}
	return &pdfService{
		generator:  generator,
		signerName: signerName,
	}
}

func (s *pdfService) AssignmentPDF(vehicle *model.Vehicle) ([]byte, error) {
	if vehicle.Status != model.StatusAssigned || vehicle.AssignedTo == nil {
		return nil, errors.New("vehicle is not currently assigned")
	}

	data := gotenberg.AssignmentData{
		SignerName:   s.signerName,
		AssigneeName: fmt.Sprintf("%s %s", vehicle.AssignedTo.FirstName, vehicle.AssignedTo.LastName),
		VehicleBrand: vehicle.Brand,
		VehicleModel: vehicle.Model,
		LicensePlate: vehicle.LicensePlate,
		IsJolly:      vehicle.Jolly,
	}

	return s.generator.GenerateAssignment(data)
}
