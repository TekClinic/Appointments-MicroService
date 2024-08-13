package main

import (
	"context"
	"time"

	ppb "github.com/TekClinic/Appointments-MicroService/appointments_protobuf"
	"github.com/uptrace/bun"
)

// Appointment defines a schema of appointments.
type Appointment struct {
	ID                int32 `bun:",pk,autoincrement"`
	PatientID         int32
	DoctorID          int32
	StartTime         time.Time
	EndTime           time.Time
	ApprovedByPatient bool
	Visited           bool
}

// toGRPC returns a GRPC version of Appointment.
func (appointment Appointment) toGRPC() *ppb.GetAppointmentResponse {
	return &ppb.GetAppointmentResponse{
		Id:                appointment.ID,
		PatientId:         appointment.PatientID,
		DoctorId:          appointment.DoctorID,
		StartTime:         appointment.StartTime.Format(time.RFC3339),
		EndTime:           appointment.EndTime.Format(time.RFC3339),
		ApprovedByPatient: appointment.ApprovedByPatient,
		Visited:           appointment.Visited,
	}
}

// createSchemaIfNotExists creates all required schemas for appointment microservice.
func createSchemaIfNotExists(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Appointment)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().IfNotExists().Model(model).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
