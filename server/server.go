package server

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

type appointmentsServer struct {
	ppb.UnimplementedAppointmentsServiceServer
	ms.BaseServiceServer
	db *bun.DB
}

const permissionDeniedMessage = "You don't have enough permission to access this resource"

// We are here


func (server appointmentsServer) GetAppointment(ctx context.Context, req *ppb.GetAppointmentRequest) (*ppb.Appointment, error) {
	
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	appointment := new(Appointment)

	err = server.db.NewSelect().Model(appointment).Where("? = ?", bun.Ident("id"), req.Id).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch an appointment by id: %w", err).Error())
	}
	if appointment == nil {
		return nil, status.Error(codes.NotFound, "Appointment is not found")
	}
	return appointment.toGRPC(), nil

}

func (server appointmentsServer) CreateAppointment(ctx context.Context, req *ppb.PostAppointmentRequest) (*ppb.AppointmentId, error) {

}

func (server appointmentsServer) GetAppointments(req *ppb.RangeRequest, dispatcher ppb.AppointmentsService_GetAppointmentsServer) error {
	
	claims, err := server.VerifyToken(dispatcher.Context(), req.GetToken())
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	// Parse date from request
	date := time.Date(int(req.GetDate().GetYear()), time.Month(req.GetDate().GetMonth()), int(req.GetDate().GetDay()), 0, 0, 0, 0, time.UTC)

	// Fetch appointments based on filters
	var appointments []Appointment
	query := server.db.NewSelect().Model(&appointments)

	// Filter by date
	query = query.Where("start_time >= ?", date).Where("end_time < ?", date.AddDate(0, 0, 1))

	// Filter by doctor ID
	if req.GetDoctorId() != "" {
		query = query.Where("doctor_id = ?", req.GetDoctorId())
	}

	// Filter by patient ID
	if req.GetPatientId() != "" {
		query = query.Where("patient_id = ?", req.GetPatientId())
	}

	// Execute query
	if err := query.Scan(dispatcher.Context()); err != nil {
		return status.Error(codes.Internal, fmt.Errorf("failed to fetch appointments: %w", err).Error())
	}

	// Send appointments to the client
	for _, appointment := range appointments {
		if err := dispatcher.Send(appointment.toGRPC()); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("failed to send appointment: %w", err).Error())
		}
	}

	return nil

}

func (server appointmentsServer) AssignPatient(ctx context.Context, req *ppb.AppointmentIdRequest) (*ppb.PatientId, error) {

}

func (server appointmentsServer) RemovePatient(ctx context.Context, req *ppb.AppointmentIdRequest) (*ppb.PatientId, error) {
}

func (server appointmentsServer) DeleteAppointment(ctx context.Context, req *ppb.AppointmentIdRequest) (*ppb.DeletedMessage, error) {
}







func (server patientsServer) GetPatients(req *ppb.RangeRequest, dispatcher ppb.PatientsService_GetPatientsServer) error {
	claims, err := server.VerifyToken(dispatcher.Context(), req.GetToken())
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}
	if req.Offset < 0 {
		return status.Error(codes.InvalidArgument, "offset has to be a non-negative integer")
	}

	if req.Limit <= 0 {
		return status.Error(codes.InvalidArgument, "offset has to be a positive integer")
	}

	var patients []Patient
	err = server.db.NewSelect().Model(&patients).Offset(int(req.Offset)).Limit(int(req.Limit)).Scan(dispatcher.Context())
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("failed to fetch users: %w", err).Error())
	}

	for _, patient := range patients {
		if err := dispatcher.Send(patient.toGRPC()); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("error occcured while sending users: %w", err).Error())
		}
	}
	return nil
}

// createPatientsServer initializing a PatientServer with all the necessary fields.
func createPatientsServer() (*patientsServer, error) {
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
		pgdriver.WithApplicationName("patients"),
		pgdriver.WithInsecure(true),
	)
	db := bun.NewDB(sql.OpenDB(connector), pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))
	return &patientsServer{BaseServiceServer: base, db: db}, nil
}

func main() {
	service, err := createPatientsServer()
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
	ppb.RegisterPatientsServiceServer(srv, service)

	log.Println("Server listening on :" + service.GetPort())
	if err := srv.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
