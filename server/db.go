package main

import (
	"context"
	"time"

	ppb "github.com/TekClinic/Appointments-MicroService/appointments_protobuf"
	sf "github.com/sa-/slicefunk"
	"github.com/uptrace/bun"
)

type Appointment struct {
	ID						int `bun:",pk,autoincrement"`
	PatientID				int
	DoctorID				int
	StartTime				time.Time
	EndTime					time.Time
	ApprovedByPatient		bool
	Visited					bool				
}


func (personalId PersonalId) toGRPC() *ppb.Patient_PersonalId {
	return &ppb.Patient_PersonalId{
		Id:   personalId.ID,
		Type: personalId.Type,
	}
}

func (contact EmergencyContact) toGRPC() *ppb.Patient_EmergencyContact {
	return &ppb.Patient_EmergencyContact{
		Name:      contact.Name,
		Closeness: contact.Closeness,
		Phone:     contact.Phone,
	}
}

func createGRPCDate(time time.Time) *ppb.Patient_Date {
	return &ppb.Patient_Date{
		Day:   int32(time.Day()),
		Month: int32(time.Month()),
		Year:  int32(time.Year()),
	}
}

func (patient Patient) toGRPC() *ppb.Patient {
	EmergencyContacts := sf.Map(patient.EmergencyContacts,
		func(contact *EmergencyContact) *ppb.Patient_EmergencyContact { return contact.toGRPC() })
	return &ppb.Patient{
		Id:                int64(patient.ID),
		Active:            patient.Active,
		Name:              patient.Name,
		PersonalId:        patient.PersonalId.toGRPC(),
		Gender:            patient.Gender,
		PhoneNumber:       patient.PhoneNumber,
		Languages:         patient.Languages,
		BirthDate:         createGRPCDate(patient.BirthDate),
		Age:               int32(time.Now().Year() - patient.BirthDate.Year()),
		ReferredBy:        patient.ReferredBy,
		EmergencyContacts: EmergencyContacts,
		SpecialNote:       patient.SpecialNote,
	}
}

func createSchemaIfNotExists(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Patient)(nil),
		(*EmergencyContact)(nil),
	}

	for _, model := range models {
		if _, err := db.NewCreateTable().IfNotExists().Model(model).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
