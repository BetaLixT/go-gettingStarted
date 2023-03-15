//go:build ignore

// Contains instructions for required generators, invoke by entering
// go generate gen.go

package main

//go:generate protoc --go_out=paths=source_relative:./pkg/domain/ proto/contracts/models.proto
// protoc is sometimes just awful
//go:generate rm -r pkg/domain/contracts
//go:generate mv pkg/domain/proto/contracts pkg/domain/
//go:generate rm -r pkg/domain/proto/
//go:generate protoc --go-grpc_out=. --gocqrshttp_out=. proto/contracts/service.proto
//go:generate hndlrgen
//go:generate protoc -I=pkg/impls/evcqrs/entities --go_out=. pkg/impls/evcqrs/entities/data.proto
//go:generate cp pkg/app/server/contracts/service.http.json pkg/app/server/static/swagger/swagger.json
//go:generate protoc --go_out=paths=source_relative:./contracts/pkg/domain/ proto/contracts/models.proto
//go:generate rm -r contracts/pkg/domain/contracts
//go:generate mv contracts/pkg/domain/proto/contracts contracts/pkg/domain/
//go:generate rm -r contracts/pkg/domain/proto/
//go:generate protoc --go-grpc_out=./contracts  proto/contracts/service.proto

//go:generate wire ./pkg/app/server
