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

// Appointment defines a schema of appointments.
type Appointment struct {
	ID                int `bun:",pk,autoincrement"`
	PatientID         int
	DoctorID          int
	StartTime         time.Time
	EndTime           time.Time
	ApprovedByPatient bool
	Visited           bool
}

// toGRPC returns a GRPC version of Appointment.
func (appointment Appointment) toGRPC() *ppb.Appointment {
	return &ppb.Appointment{
		Id:                int64(appointment.ID),
		PatientId:         int64(appointment.PatientID),
		DoctorId:          int64(appointment.DoctorID),
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

// appointmentsServer is an implementation of GRPC appointment ms. It provides access to database via db field.
type appointmentsServer struct {
	ppb.UnimplementedAppointmentsServiceServer
	ms.BaseServiceServer
	db *bun.DB
}

const permissionDeniedMessage = "You don't have enough permission to access this resource"

// GetAppointment returns the appointment information corresponding to the given ID.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If the appointment with the given ID doesn't exist, codes.NotFound is returned.
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
	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), req.GetId()).Scan(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	return appointment.toGRPC(), nil
}

// CreateAppointment creates a new appointment based on the provided details.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in parsing the start or end time, an appropriate error is returned.
// If there's an error in creating the appointment, an appropriate error is returned.
func (server appointmentsServer) CreateAppointment(
	ctx context.Context,
	req *ppb.PostAppointmentRequest,
) (*ppb.AppointmentID, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	// Assuming req.GetStartTime() and req.GetEndTime() return strings in "2006-01-02T15:04:05Z" format
	startTimeStr := req.GetStartTime()
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse start time: %w", err).Error())
	}

	endTimeStr := req.GetEndTime()
	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse end time: %w", err).Error())
	}

	appointment := Appointment{
		PatientID:         int(req.GetPatientId()),
		DoctorID:          int(req.GetDoctorId()),
		StartTime:         startTime,
		EndTime:           endTime,
		ApprovedByPatient: false,
		Visited:           false,
	}

	_, err = server.db.NewInsert().Model(&appointment).Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to create an appointment: %w", err).Error())
	}

	return &ppb.AppointmentID{Id: int64(appointment.ID)}, nil
}

// GetAppointments returns a list of appointments based on provided filters.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in parsing the date or fetching appointments, an appropriate error is returned.
func (server appointmentsServer) GetAppointments(ctx context.Context,
	req *ppb.RangeRequest) (*ppb.AppointmentsResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	// Assuming req.GetDate() returns a string in "2006-01-02" format
	dateStr := req.GetDate()
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse date: %w", err).Error())
	}

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

// AssignPatient assigns a patient to an existing appointment.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or updating the appointment, an appropriate error is returned.
func (server appointmentsServer) AssignPatient(ctx context.Context,
	req *ppb.AssignPatientRequest) (*ppb.PatientID, error) {
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

	return &ppb.PatientID{PatientId: req.GetPatientId()}, nil
}

// RemovePatient removes a patient from an existing appointment.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or updating the appointment, an appropriate error is returned.
func (server appointmentsServer) RemovePatient(ctx context.Context,
	req *ppb.AppointmentIDRequest) (*ppb.PatientID, error) {
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

	return &ppb.PatientID{PatientId: int64(patientID)}, nil
}

// DeleteAppointment deletes an appointment based on the provided ID.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or deleting the appointment, an appropriate error is returned.
func (server appointmentsServer) DeleteAppointment(ctx context.Context,
	req *ppb.AppointmentIDRequest) (*ppb.DeletedMessage, error) {
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
