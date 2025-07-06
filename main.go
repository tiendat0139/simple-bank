package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tiendat0139/simple-bank/api"
	db "github.com/tiendat0139/simple-bank/db/sqlc"
	"github.com/tiendat0139/simple-bank/gapi"
	"github.com/tiendat0139/simple-bank/pb"
	"github.com/tiendat0139/simple-bank/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msg("cannot load config:")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Msg("cannot connect to db:")
	}

	runDBMigration(config.MigrationURL, config.DBSource)

	store := db.NewStore(conn)
	go runGatewayServer(config, store)
	runGrpcServer(config, store)

}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Msg("cannot create migration instance:")
	}

	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msg("failed to run migration:")
	}

	log.Info().Msg("DB migration completed successfully")
}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener:")
	}
	log.Info().Msgf("starting gRPC server at %s\n", config.GRPCServerAddress)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Msg("cannot start gRPC server:")
	}

}

func runGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	// By default, the gRPC compliler generates camelCase field names,
	// Use code below to use snake_case field names in the JSON response.
	jsonOptions := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOptions)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msg("cannot register gRPC gateway:")
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msg("cannot create listener:")
	}

	log.Info().Msgf("starting HTTP gateway server at %s\n", listener.Addr().String())
	httpHandler := gapi.HttpLogger(mux)
	err = http.Serve(listener, httpHandler)
	if err != nil {
		log.Fatal().Msg("cannot start gRPC server:")
	}

}

func runGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Msg("cannot create server:")
	}

	err = server.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Msg("cannot start server:")
	}
}
