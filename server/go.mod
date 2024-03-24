module github.com/TekClinic/Appointments-MicroService/server

go 1.22.0

require (
	github.com/TekClinic/Appointments-MicroService/appointments_protobuf v0.1.1
	github.com/TekClinic/MicroService-Lib v0.1.1
	github.com/uptrace/bun v1.1.17
	github.com/uptrace/bun/dialect/pgdialect v1.1.17
	github.com/uptrace/bun/driver/pgdriver v1.1.17
	github.com/uptrace/bun/extra/bundebug v1.1.17
	google.golang.org/grpc v1.62.1
)

require (
	github.com/coreos/go-oidc/v3 v3.9.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/sa-/slicefunk v0.1.4 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/oauth2 v0.16.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	k8s.io/apimachinery v0.29.2 // indirect
	mellium.im/sasl v0.3.1 // indirect
)

replace github.com/TekClinic/Appointments-MicroService/appointments_protobuf => ./../appointments_protobuf
