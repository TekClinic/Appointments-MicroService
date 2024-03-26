// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: appointments_service.proto

package appointments_protobuf

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AppointmentsServiceClient is the client API for AppointmentsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AppointmentsServiceClient interface {
	GetAppointment(ctx context.Context, in *GetAppointmentRequest, opts ...grpc.CallOption) (*Appointment, error)
	CreateAppointment(ctx context.Context, in *PostAppointmentRequest, opts ...grpc.CallOption) (*AppointmentId, error)
	GetAppointments(ctx context.Context, in *RangeRequest, opts ...grpc.CallOption) (*AppointmentsResponse, error)
	AssignPatient(ctx context.Context, in *AssignPatientRequest, opts ...grpc.CallOption) (*PatientId, error)
	RemovePatient(ctx context.Context, in *AppointmentIdRequest, opts ...grpc.CallOption) (*PatientId, error)
	DeleteAppointment(ctx context.Context, in *AppointmentIdRequest, opts ...grpc.CallOption) (*DeletedMessage, error)
}

type appointmentsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAppointmentsServiceClient(cc grpc.ClientConnInterface) AppointmentsServiceClient {
	return &appointmentsServiceClient{cc}
}

func (c *appointmentsServiceClient) GetAppointment(ctx context.Context, in *GetAppointmentRequest, opts ...grpc.CallOption) (*Appointment, error) {
	out := new(Appointment)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/GetAppointment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appointmentsServiceClient) CreateAppointment(ctx context.Context, in *PostAppointmentRequest, opts ...grpc.CallOption) (*AppointmentId, error) {
	out := new(AppointmentId)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/CreateAppointment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appointmentsServiceClient) GetAppointments(ctx context.Context, in *RangeRequest, opts ...grpc.CallOption) (*AppointmentsResponse, error) {
	out := new(AppointmentsResponse)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/GetAppointments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appointmentsServiceClient) AssignPatient(ctx context.Context, in *AssignPatientRequest, opts ...grpc.CallOption) (*PatientId, error) {
	out := new(PatientId)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/AssignPatient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appointmentsServiceClient) RemovePatient(ctx context.Context, in *AppointmentIdRequest, opts ...grpc.CallOption) (*PatientId, error) {
	out := new(PatientId)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/RemovePatient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *appointmentsServiceClient) DeleteAppointment(ctx context.Context, in *AppointmentIdRequest, opts ...grpc.CallOption) (*DeletedMessage, error) {
	out := new(DeletedMessage)
	err := c.cc.Invoke(ctx, "/appointments.AppointmentsService/DeleteAppointment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AppointmentsServiceServer is the server API for AppointmentsService service.
// All implementations must embed UnimplementedAppointmentsServiceServer
// for forward compatibility
type AppointmentsServiceServer interface {
	GetAppointment(context.Context, *GetAppointmentRequest) (*Appointment, error)
	CreateAppointment(context.Context, *PostAppointmentRequest) (*AppointmentId, error)
	GetAppointments(context.Context, *RangeRequest) (*AppointmentsResponse, error)
	AssignPatient(context.Context, *AssignPatientRequest) (*PatientId, error)
	RemovePatient(context.Context, *AppointmentIdRequest) (*PatientId, error)
	DeleteAppointment(context.Context, *AppointmentIdRequest) (*DeletedMessage, error)
	mustEmbedUnimplementedAppointmentsServiceServer()
}

// UnimplementedAppointmentsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAppointmentsServiceServer struct {
}

func (UnimplementedAppointmentsServiceServer) GetAppointment(context.Context, *GetAppointmentRequest) (*Appointment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAppointment not implemented")
}
func (UnimplementedAppointmentsServiceServer) CreateAppointment(context.Context, *PostAppointmentRequest) (*AppointmentId, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAppointment not implemented")
}
func (UnimplementedAppointmentsServiceServer) GetAppointments(context.Context, *RangeRequest) (*AppointmentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAppointments not implemented")
}
func (UnimplementedAppointmentsServiceServer) AssignPatient(context.Context, *AssignPatientRequest) (*PatientId, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AssignPatient not implemented")
}
func (UnimplementedAppointmentsServiceServer) RemovePatient(context.Context, *AppointmentIdRequest) (*PatientId, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemovePatient not implemented")
}
func (UnimplementedAppointmentsServiceServer) DeleteAppointment(context.Context, *AppointmentIdRequest) (*DeletedMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAppointment not implemented")
}
func (UnimplementedAppointmentsServiceServer) mustEmbedUnimplementedAppointmentsServiceServer() {}

// UnsafeAppointmentsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AppointmentsServiceServer will
// result in compilation errors.
type UnsafeAppointmentsServiceServer interface {
	mustEmbedUnimplementedAppointmentsServiceServer()
}

func RegisterAppointmentsServiceServer(s grpc.ServiceRegistrar, srv AppointmentsServiceServer) {
	s.RegisterService(&AppointmentsService_ServiceDesc, srv)
}

func _AppointmentsService_GetAppointment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAppointmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).GetAppointment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/GetAppointment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).GetAppointment(ctx, req.(*GetAppointmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppointmentsService_CreateAppointment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostAppointmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).CreateAppointment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/CreateAppointment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).CreateAppointment(ctx, req.(*PostAppointmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppointmentsService_GetAppointments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RangeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).GetAppointments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/GetAppointments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).GetAppointments(ctx, req.(*RangeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppointmentsService_AssignPatient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AssignPatientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).AssignPatient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/AssignPatient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).AssignPatient(ctx, req.(*AssignPatientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppointmentsService_RemovePatient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppointmentIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).RemovePatient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/RemovePatient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).RemovePatient(ctx, req.(*AppointmentIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AppointmentsService_DeleteAppointment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AppointmentIdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppointmentsServiceServer).DeleteAppointment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/appointments.AppointmentsService/DeleteAppointment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppointmentsServiceServer).DeleteAppointment(ctx, req.(*AppointmentIdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AppointmentsService_ServiceDesc is the grpc.ServiceDesc for AppointmentsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AppointmentsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "appointments.AppointmentsService",
	HandlerType: (*AppointmentsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAppointment",
			Handler:    _AppointmentsService_GetAppointment_Handler,
		},
		{
			MethodName: "CreateAppointment",
			Handler:    _AppointmentsService_CreateAppointment_Handler,
		},
		{
			MethodName: "GetAppointments",
			Handler:    _AppointmentsService_GetAppointments_Handler,
		},
		{
			MethodName: "AssignPatient",
			Handler:    _AppointmentsService_AssignPatient_Handler,
		},
		{
			MethodName: "RemovePatient",
			Handler:    _AppointmentsService_RemovePatient_Handler,
		},
		{
			MethodName: "DeleteAppointment",
			Handler:    _AppointmentsService_DeleteAppointment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "appointments_service.proto",
}
