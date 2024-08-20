## gRPC Functions

### GetAppointment

Retrieves the details of a specific appointment by its ID.

**Request:**

```protobuf
message GetAppointmentRequest {
  string token = 1; // Authentication token
  int32 id = 2; // ID of the appointment
}
```

**Response:**

```protobuf
message GetAppointmentResponse {
  int32 id = 1; // ID of the appointment
  int32 patient_id = 2; // ID of the assigned patient
  int32 doctor_id = 3; // ID of the assigned doctor
  string start_time = 4; // Start time of the appointment
  string end_time = 5; // End time of the appointment
  bool approved_by_patient = 6; // Whether the appointment is approved by the patient
  bool visited = 7; // Whether the patient has visited
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Appointment with the given ID does not exist.

---

### CreateAppointment

Creates a new appointment with the provided details.

**Request:**

```protobuf
message CreateAppointmentRequest {
  string token = 1; // Authentication token
  int32 patient_id = 2; // ID of the patient
  int32 doctor_id = 3; // ID of the doctor
  string start_time = 4; // Start time of the appointment
  string end_time = 5; // End time of the appointment
}
```

**Response:**

```protobuf
message CreateAppointmentResponse {
  int32 id = 1; // ID of the newly created appointment
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - Required appointment information is missing or malformed.

---

### GetAppointments

Retrieves a list of appointments with optional filtering by date, doctor, and patient, with pagination support.

**Request:**

```protobuf
message GetAppointmentsRequest {
  string token = 1; // Authentication token
  string date = 2; // Date to filter appointments (optional)
  int32 doctor_id = 3; // ID of the doctor to filter appointments (optional)
  int32 patient_id = 4; // ID of the patient to filter appointments (optional)
  int32 skip = 5; // Number of appointments to skip (for pagination)
  int32 limit = 6; // Maximum number of appointments to return
}
```

**Response:**

```protobuf
message GetAppointmentsResponse {
  int32 count = 1; // Total number of appointments matching the filters
  repeated int32 results = 2; // List of appointment IDs
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - `skip`, `limit`, or filter parameters are invalid.

---

### AssignPatient

Assigns a patient to an existing appointment.

**Request:**

```protobuf
message AssignPatientRequest {
  string token = 1; // Authentication token
  int32 id = 2; // ID of the appointment
  int32 patient_id = 3; // ID of the patient to be assigned
}
```

**Response:**

```protobuf
message AssignPatientResponse {
  int32 patient_id = 1; // ID of the assigned patient
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Appointment with the given ID does not exist.

---

### RemovePatient

Removes a patient from an existing appointment.

**Request:**

```protobuf
message RemovePatientRequest {
  string token = 1; // Authentication token
  int32 id = 2; // ID of the appointment
}
```

**Response:**

```protobuf
message RemovePatientResponse {
  int32 patient_id = 1; // ID of the removed patient
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Appointment with the given ID does not exist.

---

### DeleteAppointment

Deletes an appointment by its ID.

**Request:**

```protobuf
message DeleteAppointmentRequest {
  string token = 1; // Authentication token
  int32 id = 2; // ID of the appointment to be deleted
}
```

**Response:**

```protobuf
message DeleteAppointmentResponse {
  string message = 1; // Confirmation message
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `NotFound` - Appointment with the given ID does not exist.

---

### UpdateAppointment

Updates the details of an existing appointment.

**Request:**

```protobuf
message UpdateAppointmentRequest {
  string token = 1; // Authentication token
  int32 id = 2; // ID of the appointment to be updated
  int32 patient_id = 3; // ID of the patient
  int32 doctor_id = 4; // ID of the doctor
  string start_time = 5; // Start time of the appointment
  string end_time = 6; // End time of the appointment
  bool approved_by_patient = 7; // Whether the appointment is approved by the patient
  bool visited = 8; // Whether the patient has visited
}
```

**Response:**

```protobuf
message UpdateAppointmentResponse {
  int32 id = 1; // ID of the updated appointment
}
```

**Errors:**

- `Unauthenticated` - Token is not valid or expired.
- `PermissionDenied` - Token is not authorized with the *admin* role.
- `InvalidArgument` - Updated appointment information is missing or malformed.
- `NotFound` - Appointment with the given ID does not exist.

---

## Model Definition

```protobuf
message Appointment {
  int32 id = 1; // ID of the appointment
  int32 patient_id = 2; // ID of the assigned patient
  int32 doctor_id = 3; // ID of the assigned doctor
  string start_time = 4; // Start time of the appointment
  string end_time = 5; // End time of the appointment
  bool approved_by_patient = 6; // Whether the appointment is approved by the patient
  bool visited = 7; // Whether the patient has visited
}
```