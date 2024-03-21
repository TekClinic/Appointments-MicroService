package main

import (
	"context"
	"time"

	ppb "github.com/TekClinic/Appointments-MicroService/appointments_protobuf"
	"github.com/uptrace/bun"
)

type Appointment struct {
	ID						int `bun:",pk,autoincrement"`
	PatientId				int
	DoctorId				int
	StartTime				time.Time
	EndTime					time.Time
	ApprovedByPatient		bool
	Visited					bool				
}

func createGRPCDate(time time.Time) *ppb.Time {
	return &ppb.Time{
		Day:   int32(time.Day()),
		Month: int32(time.Month()),
		Year:  int32(time.Year()),
		Hour:	int32(time.Hour()),
		Minute: int32(time.Minute()),
	}
}

func (appointment Appointment) toGRPC() *ppb.Appointment {
	return &ppb.Appointment{
		Id:              	    int64(appointment.ID),
		PatientId:        	 	int64(appointment.PatientId),
		DoctorId:               int64(appointment.DoctorId),
		StartTime:         	    createGRPCDate(appointment.StartTime),
		EndTime:       			createGRPCDate(appointment.EndTime),
		ApprovedByPatient:      appointment.ApprovedByPatient,
		Visited: 				appointment.Visited,
	}
}

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
