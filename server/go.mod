module github.com/TekClinic/Appointments-MicroService/server

go 1.22.0

require (
	github.com/TekClinic/Appointments-MicroService/appointments_protobuf v0.100.0-integrated
	github.com/TekClinic/MicroService-Lib v0.1.3
	github.com/uptrace/bun v1.2.1
	github.com/uptrace/bun/dialect/pgdialect v1.2.1
	github.com/uptrace/bun/driver/pgdriver v1.2.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.65.0
)

require (
	github.com/alexlast/bunzap v0.1.0 // indirect
	github.com/coreos/go-oidc/v3 v3.11.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/sa-/slicefunk v0.1.4 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.26.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240812133136-8ffd90a71988 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	k8s.io/apimachinery v0.31.0 // indirect
	mellium.im/sasl v0.3.1 // indirect
)

replace github.com/TekClinic/Appointments-MicroService/appointments_protobuf => ./../appointments_protobuf
