package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	ppb "github.com/TekClinic/Appointments-MicroService/appointments_protobuf"

	ms "github.com/TekClinic/MicroService-Lib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Appointment struct {
	ID                int `bun:",pk,autoincrement"`
	PatientID         int
	DoctorID          int
	StartTime         time.Time
	EndTime           time.Time
	ApprovedByPatient bool
	Visited           bool
}

func createGRPCDate(time time.Time) *ppb.Time {
	return &ppb.Time{
		Day:    int32(time.Day()),
		Month:  int32(time.Month()),
		Year:   int32(time.Year()),
		Hour:   int32(time.Hour()),
		Minute: int32(time.Minute()),
	}
}

func (appointment Appointment) toGRPC() *ppb.Appointment {
	return &ppb.Appointment{
		Id:                int64(appointment.ID),
		PatientId:         int64(appointment.PatientID),
		DoctorId:          int64(appointment.DoctorID),
		StartTime:         createGRPCDate(appointment.StartTime),
		EndTime:           createGRPCDate(appointment.EndTime),
		ApprovedByPatient: appointment.ApprovedByPatient,
		Visited:           appointment.Visited,
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

type appointmentsServer struct {
	ppb.UnimplementedAppointmentsServiceServer
	ms.BaseServiceServer
	db *bun.DB
}

const permissionDeniedMessage = "You don't have enough permission to access this resource"

func (server appointmentsServer) GetAppointment(ctx context.Context,
	req *ppb.GetAppointmentRequest) (*ppb.Appointment, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().Model(appointment).Where("id = ?", req.GetId()).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	return appointment.toGRPC(), nil
}

func (server appointmentsServer) CreateAppointment(ctx context.Context,
	req *ppb.PostAppointmentRequest) (*ppb.AppointmentId, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := Appointment{
		PatientID: int(req.GetPatientId()),
		DoctorID:  int(req.GetDoctorId()),
		StartTime: time.Date(
			int(req.GetStartTime().GetYear()),
			time.Month(req.GetStartTime().GetMonth()),
			int(req.GetStartTime().GetDay()),
			int(req.GetStartTime().GetHour()),
			int(req.GetStartTime().GetMinute()),
			0, 0, time.UTC,
		),
		EndTime: time.Date(
			int(req.GetEndTime().GetYear()),
			time.Month(req.GetEndTime().GetMonth()),
			int(req.GetEndTime().GetDay()),
			int(req.GetEndTime().GetHour()),
			int(req.GetEndTime().GetMinute()),
			0, 0, time.UTC,
		),
		ApprovedByPatient: false,
		Visited:           false,
	}

	_, err = server.db.NewInsert().Model(&appointment).Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to create an appointment: %w", err).Error())
	}

	return &ppb.AppointmentId{Id: int64(appointment.ID)}, nil
}

func (server appointmentsServer) GetAppointments(ctx context.Context,
	req *ppb.RangeRequest) (*ppb.AppointmentsResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	// Parse date from request
	date := time.Date(int(req.GetDate().GetYear()),
		time.Month(req.GetDate().GetMonth()), int(req.GetDate().GetDay()), 0, 0, 0, 0, time.UTC)

	// Fetch appointments based on filters
	baseQuery := server.db.NewSelect().Model((*Appointment)(nil))

	// Filter by date range
	baseQuery = baseQuery.Where("start_time >= ?", date).Where("end_time <= ?", date.AddDate(0, 0, 1))

	// Filter by doctor ID
	if req.GetDoctorId() != "" {
		baseQuery = baseQuery.Where("doctor_id = ?", req.GetDoctorId())
	}

	// Filter by patient ID
	if req.GetPatientId() != "" {
		baseQuery = baseQuery.Where("patient_id = ?", req.GetPatientId())
	}

	// Execute query and get count
	var ids []int32
	err = baseQuery.Column("id").Scan(ctx, &ids)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch appointment IDs: %w", err).Error())
	}

	count, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to count appointments: %w", err).Error())
	}

	return &ppb.AppointmentsResponse{
		Count:   int32(count),
		Results: ids,
	}, nil
}

func (server appointmentsServer) AssignPatient(ctx context.Context,
	req *ppb.AssignPatientRequest) (*ppb.PatientId, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), req.GetAppointmentId()).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	appointment.PatientID = int(req.GetPatientId())
	_, err = server.db.NewUpdate().Model(appointment).WherePK().Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to assign patient to appointment: %w", err).Error())
	}

	return &ppb.PatientId{PatientId: req.GetPatientId()}, nil
}

func (server appointmentsServer) RemovePatient(ctx context.Context,
	req *ppb.AppointmentIdRequest) (*ppb.PatientId, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), req.GetId()).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	patientID := appointment.PatientID
	appointment.PatientID = 0
	_, err = server.db.NewUpdate().Model(appointment).WherePK().Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to remove patient from appointment: %w", err).Error())
	}

	return &ppb.PatientId{PatientId: int64(patientID)}, nil
}

func (server appointmentsServer) DeleteAppointment(ctx context.Context,
	req *ppb.AppointmentIdRequest) (*ppb.DeletedMessage, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), req.GetId()).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	_, err = server.db.NewDelete().Model(appointment).WherePK().Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to delete appointment: %w", err).Error())
	}

	return &ppb.DeletedMessage{Messgae: "Appointment deleted successfully"}, nil
}

// createAppointmentsServer initializing an AppointmentServer with all the necessary fields.
func createAppointmentsServer() (*appointmentsServer, error) {
	base, err := ms.CreateBaseServiceServer()
	if err != nil {
		return nil, err
	}
	addr, err := ms.GetRequiredEnv("DB_ADDR")
	if err != nil {
		return nil, err
	}
	user, err := ms.GetRequiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}
	password, err := ms.GetRequiredEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}
	database, err := ms.GetRequiredEnv("DB_DATABASE")
	if err != nil {
		return nil, err
	}
	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(addr),
		pgdriver.WithUser(user),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(database),
		pgdriver.WithApplicationName("appointments"),
		pgdriver.WithInsecure(true),
	)
	db := bun.NewDB(sql.OpenDB(connector), pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return &appointmentsServer{BaseServiceServer: base, db: db}, nil
}

func main() {
	service, err := createAppointmentsServer()
	if err != nil {
		log.Fatal(err)
	}

	err = createSchemaIfNotExists(context.Background(), service.db)
	if err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("tcp", ":"+service.GetPort())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	ppb.RegisterAppointmentsServiceServer(srv, service)

	log.Println("Server listening on :" + service.GetPort())
	if err = srv.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
