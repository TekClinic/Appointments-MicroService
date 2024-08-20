package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	ppb "github.com/TekClinic/Appointments-MicroService/appointments_protobuf"

	ms "github.com/TekClinic/MicroService-Lib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// appointmentsServer is an implementation of GRPC appointment ms. It provides access to a database via db field.
type appointmentsServer struct {
	ppb.UnimplementedAppointmentsServiceServer
	ms.BaseServiceServer
	db *bun.DB
}

const (
	envDBAddress  = "DB_ADDR"
	envDBUser     = "DB_USER"
	envDBDatabase = "DB_DATABASE"
	envDBPassword = "DB_PASSWORD"

	applicationName = "appointments"

	permissionDeniedMessage = "You don't have enough permission to access this resource"

	maxPaginationLimit = 50
	dateFormat         = "2006-01-02"
)

// GetAppointment returns the appointment information corresponding to the given ID.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If the appointment with the given ID doesn't exist, codes.NotFound is returned.
func (server appointmentsServer) GetAppointment(ctx context.Context,
	req *ppb.GetAppointmentRequest) (*ppb.GetAppointmentResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().
		Model(appointment).
		Where("? = ?", bun.Ident("id"), req.GetId()).
		WhereAllWithDeleted().
		Scan(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	return appointment.toGRPC(), nil
}

// CreateAppointment creates a new appointment based on the provided details.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in parsing the start or end time, an appropriate error is returned.
// If there's an error in creating the appointment, an appropriate error is returned.
func (server appointmentsServer) CreateAppointment(
	ctx context.Context,
	req *ppb.CreateAppointmentRequest,
) (*ppb.CreateAppointmentResponse, error) {
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

	patientID := req.GetPatientId()
	doctorID := req.GetDoctorId()
	if doctorID == 0 {
		return nil, status.Error(codes.InvalidArgument,
			errors.New("DoctorID is required in order to create an appointment").Error())
	}
	if patientID < 0 || doctorID < 0 {
		return nil, status.Error(codes.InvalidArgument,
			errors.New("PatientID, DoctorID have to be non-negative values").Error())
	}

	appointment := Appointment{
		PatientID:         patientID,
		DoctorID:          doctorID,
		StartTime:         startTime,
		EndTime:           endTime,
		ApprovedByPatient: false,
		Visited:           false,
	}

	_, err = server.db.NewInsert().Model(&appointment).Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to create an appointment: %w", err).Error())
	}

	return &ppb.CreateAppointmentResponse{Id: appointment.ID}, nil
}

// GetAppointments returns a list of appointments based on provided filters.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in parsing the date or fetching appointments, an appropriate error is returned.
func (server appointmentsServer) GetAppointments(ctx context.Context,
	req *ppb.GetAppointmentsRequest) (*ppb.GetAppointmentsResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	if req.GetSkip() < 0 {
		return nil, status.Error(codes.InvalidArgument, "skip has to be a non-negative integer")
	}
	if req.GetLimit() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "limit has to be a positive integer")
	}
	if req.GetLimit() > maxPaginationLimit {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("maximum allowed limit values is %d", maxPaginationLimit))
	}

	// Fetch appointments based on filters
	baseQuery := server.db.NewSelect().Model((*Appointment)(nil))

	// Filter by date range
	dateStr := req.GetDate()
	if dateStr != "" {
		date, dateErr := time.Parse(dateFormat, dateStr)
		if dateErr != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse date: %w", dateErr).Error())
		}
		baseQuery = baseQuery.Where("start_time >= ?", date).Where("end_time <= ?", date.AddDate(0, 0, 1))
	}

	// Filter by doctor ID
	if req.GetDoctorId() != 0 {
		baseQuery = baseQuery.Where("doctor_id = ?", req.GetDoctorId())
	}

	// Filter by patient ID
	if req.GetPatientId() != 0 {
		baseQuery = baseQuery.Where("patient_id = ?", req.GetPatientId())
	}

	// Execute a query and get count
	var ids []int32
	err = baseQuery.Column("id").Offset(int(req.GetSkip())).Limit(int(req.GetLimit())).Scan(ctx, &ids)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch appointment IDs: %w", err).Error())
	}

	count, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to count appointments: %w", err).Error())
	}

	return &ppb.GetAppointmentsResponse{
		Count:   int32(count),
		Results: ids,
	}, nil
}

// AssignPatient assigns a patient to an existing appointment.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or updating the appointment, an appropriate error is returned.
func (server appointmentsServer) AssignPatient(ctx context.Context,
	req *ppb.AssignPatientRequest) (*ppb.AssignPatientResponse, error) {
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

	patientID := req.GetPatientId()
	if patientID < 0 {
		return nil, status.Error(codes.InvalidArgument,
			errors.New("PatientID has to be non-negative values").Error())
	}
	formerPatientID := appointment.PatientID
	appointment.PatientID = patientID
	_, err = server.db.NewUpdate().Model(appointment).WherePK().Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to assign patient to appointment: %w", err).Error())
	}

	return &ppb.AssignPatientResponse{PatientId: formerPatientID}, nil
}

// RemovePatient removes a patient from an existing appointment.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or updating the appointment, an appropriate error is returned.
func (server appointmentsServer) RemovePatient(ctx context.Context,
	req *ppb.RemovePatientRequest) (*ppb.RemovePatientResponse, error) {
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

	return &ppb.RemovePatientResponse{PatientId: patientID}, nil
}

// DeleteAppointment deletes an appointment based on the provided ID.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If there's an error in fetching or deleting the appointment, an appropriate error is returned.
func (server appointmentsServer) DeleteAppointment(ctx context.Context,
	req *ppb.DeleteAppointmentRequest) (*ppb.DeleteAppointmentResponse, error) {
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

	return &ppb.DeleteAppointmentResponse{Message: "Appointment deleted successfully"}, nil
}

// createAppointmentsServer initializing an AppointmentServer with all the necessary fields.
func createAppointmentsServer() (*appointmentsServer, error) {
	base, err := ms.CreateBaseServiceServer()
	if err != nil {
		return nil, err
	}
	addr, err := ms.GetRequiredEnv(envDBAddress)
	if err != nil {
		return nil, err
	}
	user, err := ms.GetRequiredEnv(envDBUser)
	if err != nil {
		return nil, err
	}
	password, err := ms.GetRequiredEnv(envDBPassword)
	if err != nil {
		return nil, err
	}
	database, err := ms.GetRequiredEnv(envDBDatabase)
	if err != nil {
		return nil, err
	}
	connector := pgdriver.NewConnector(
		pgdriver.WithNetwork("tcp"),
		pgdriver.WithAddr(addr),
		pgdriver.WithUser(user),
		pgdriver.WithPassword(password),
		pgdriver.WithDatabase(database),
		pgdriver.WithApplicationName(applicationName),
		pgdriver.WithInsecure(!ms.HasSecureConnection()),
	)
	db := bun.NewDB(sql.OpenDB(connector), pgdialect.New())
	db.AddQueryHook(ms.GetDBQueryHook())
	return &appointmentsServer{BaseServiceServer: base, db: db}, nil
}

// UpdateAppointment updates an existing appointment based on the provided details.
// Requires authentication. If authentication is not valid, codes.Unauthenticated is returned.
// Requires an admin role. If roles are not sufficient, codes.PermissionDenied is returned.
// If one of the fields has an invalid value, an appropriate error is returned.
// If the appointment with the given ID doesn't exist, codes.NotFound is returned.
// If there's an error in fetching or updating the appointment, an appropriate error is returned.
func (server appointmentsServer) UpdateAppointment(ctx context.Context,
	req *ppb.UpdateAppointmentRequest) (*ppb.UpdateAppointmentResponse, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointmentID := req.GetId()
	if appointmentID == 0 {
		return nil, status.Error(codes.InvalidArgument, "AppointmentID is required")
	}

	appointment := new(Appointment)
	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), appointmentID).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}

	patientID := req.GetPatientId()
	doctorID := req.GetDoctorId()
	if doctorID == 0 {
		return nil, status.Error(codes.InvalidArgument,
			errors.New("DoctorID is required in order to update an appointment").Error())
	}
	if patientID < 0 || doctorID < 0 {
		return nil, status.Error(codes.InvalidArgument,
			errors.New("PatientID, DoctorID have to be non-negative values").Error())
	}

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

	appointment.PatientID = patientID
	appointment.DoctorID = doctorID
	appointment.StartTime = startTime
	appointment.EndTime = endTime
	appointment.ApprovedByPatient = req.GetApprovedByPatient()
	appointment.Visited = req.GetVisited()

	_, err = server.db.NewUpdate().
		Model(appointment).
		WherePK().
		ExcludeColumn("created_at", "deleted_at").
		Exec(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to update appointment: %w", err).Error())
	}

	return &ppb.UpdateAppointmentResponse{Id: appointment.ID}, nil
}

func main() {
	service, err := createAppointmentsServer()
	if err != nil {
		zap.L().Fatal("Failed to create appointments server", zap.Error(err))
	}

	err = createSchemaIfNotExists(context.Background(), service.db)
	if err != nil {
		zap.L().Fatal("Failed to create schema", zap.Error(err))
	}

	listen, err := net.Listen("tcp", ":"+service.GetPort())
	if err != nil {
		zap.L().Fatal("Failed to listen", zap.Error(err))
	}

	srv := grpc.NewServer(ms.GetGRPCServerOptions()...)
	ppb.RegisterAppointmentsServiceServer(srv, service)

	zap.L().Info("Server listening on :" + service.GetPort())
	if err = srv.Serve(listen); err != nil {
		zap.L().Fatal("Failed to serve", zap.Error(err))
	}
}
