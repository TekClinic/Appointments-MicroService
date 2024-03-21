package main

import (
	"context"
	"database/sql"
	"fmt"
	ms "github.com/TekClinic/MicroService-Lib"
	ppb "github.com/TekClinic/Patients-MicroService/patients_protobuf"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

type patientsServer struct {
	ppb.UnimplementedPatientsServiceServer
	ms.BaseServiceServer
	db *bun.DB
}

const permissionDeniedMessage = "You don't have enough permission to access this resource"

func (server patientsServer) GetPatient(ctx context.Context, req *ppb.PatientRequest) (*ppb.Patient, error) {
	claims, err := server.VerifyToken(ctx, req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if !claims.HasRole("admin") {
		return nil, status.Error(codes.PermissionDenied, permissionDeniedMessage)
	}

	patient := new(Patient)
	err = server.db.NewSelect().Model(patient).Where("? = ?", bun.Ident("id"), req.Id).Scan(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to fetch a user by id: %w", err).Error())
	}
	if patient == nil {
		return nil, status.Error(codes.NotFound, "User is not found")
	}
	return patient.toGRPC(), nil
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
