# Appointments-MicroService

This repository contains a gRPC service for managing Tekclinic appointments. The service is implemented in Go and uses PostgreSQL as the database.

Please note that the provided code assumes the existence of a `TekClinic/MicroService-Lib` library for authentication and environment variable handling, and setting up the environment variables found in 
`TekClinic/MicroService-Lib` is a prerequisite.

## Table of Contents

- [Installation](#installation)
- [gRPC Functions](#grpc-functions)
  - [GetAppointment](#getappointment)
  - [CreateAppointment](#createappointment)
  - [GetAppointments](#getappointments)
  - [AssignPatient](#assignpatient)
  - [RemovePatient](#removepatient)
  - [DeleteAppointment](#deleteappointment)
- [Current Behavior](#current-behavior)

## Installation

1. Clone the repository:

```bash
git clone https://github.com/TekClinic/Appointments-MicroService.git
```

2. Set up the required environment variables for the PostgreSQL database:

```
DB_ADDR=<database_address>
DB_USER=<database_user>
DB_PASSWORD=<database_password>
DB_DATABASE=<database_name>
```

3. Run the server:

```bash
go run server.go
```

## gRPC Functions

### GetAppointment

Retrieves the appointment information corresponding to the given ID.

**Request:**

```protobuf
message GetAppointmentRequest {
  string token = 1; // Authentication token
  int64 id = 2; // Appointment ID
}
```

**Response:**

```protobuf
message GetAppointmentResponse {
  int64 id = 1; // Unique identifier for the appointment
  int64 patient_id = 2; // ID of the patient associated with the appointment
  int64 doctor_id = 3; // ID of the doctor associated with the appointment
  string start_time = 4; // Start time of the appointment
  string end_time = 5; // End time of the appointment
  bool approved_by_patient = 6; // Flag indicating if the appointment is approved by the patient
  bool visited = 7; // Flag indicating if the patient has visited for the appointment
}
```

### CreateAppointment

Creates a new appointment based on the provided details.

**Request:**

```protobuf
message CreateAppointmentRequest {
  string token = 1; // Authentication token
  int32 patient_id = 2; // ID of the patient associated with the appointment
  int32 doctor_id = 3; // ID of the doctor associated with the appointment
  string start_time = 4; // Format: "2006-01-02T15:04:05Z"
  string end_time = 5; // Format: "2006-01-02T15:04:05Z"
}
```

**Response:**

```protobuf
message CreateAppointmentResponse {
  int32 id = 1; // ID of the created appointment
}
```

### GetAppointments

Returns a list of appointments based on provided filters.

**Request:**

```protobuf
message GetAppointmentsRequest {
  string token = 1; // Authentication token
  string date = 2; // Format: "2006-01-02"
  int32 doctor_id = 3; // Optional filter for doctor ID
  int32 patient_id = 4; // Optional filter for patient ID
  int32 skip = 5; // Number of results to skip
  int32 limit = 6; // Maximum number of results to return
}
```

**Response:**

```protobuf
message GetAppointmentsResponse {
  int32 count = 1; // Total number of appointments
  repeated int32 results = 2; // List of appointment IDs
}
```

### AssignPatient

Assigns a patient to an existing appointment.

**Request:**

```protobuf
message AssignPatientRequest {
  string token = 1; // Authentication token
  int32 appointment_id = 2; // ID of the appointment
  int32 patient_id = 3; // ID of the patient to be assigned
}
```

**Response:**

```protobuf
message AssignPatientResponse {
  int32 patient_id = 1; // ID of the assigned patient
}
```

### RemovePatient

Removes a patient from an existing appointment.

**Request:**

```protobuf
message RemovePatientRequest {
  string token = 1; // Authentication token
  int32 appointment_id = 2; // Appointment ID
}
```

**Response:**

```protobuf
message RemovePatientResponse {
  int32 patient_id = 1; // ID of the removed patient
}
```

### DeleteAppointment

Deletes an appointment based on the provided ID.

**Request:**

```protobuf
message DeleteAppointmentRequest {
  string token = 1; // Authentication token
  int32 appointment_id = 2; // Appointment ID
}
```

**Response:**

```protobuf
message DeleteAppointmentResponse {
  string message = 1; // Success message "Appointment deleted successfully"
}
```

## Current Behavior

- An appointment with the same information as an existing appointment can be created – they would only differ in their unique IDs (`CreateAppointment`).
- IDs are always set incrementally. For example, if 5 new appointments are created and then deleted, the next created appointment's ID would be 6, although 1, 2, 3, 4, 5 are available (`CreateAppointment`, `DeleteAppointment`).
- Upon creating a new appointment, `patient_id` is optional, and if not provided is set to 0 (`CreateAppointment`).
- `AssignPatient` can override an existing patient for a given appointment. For example, if appointment number 1 already exists with patient 1, calling `AssignPatient` with appointment number 1 and patient number 2 replaces the patient.
- After using `RemovePatient` on a given appointment, its `patient_id` field is set to 0.
- `patient_id` and `doctor_id` values are made sure to be non-negative (`CreateAppointment`, `AssignPatient`).