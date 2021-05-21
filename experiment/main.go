package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/grpc-ecosystem/grpc-gateway/v2/experiment/proto/api"
	gwpb "github.com/grpc-ecosystem/grpc-gateway/v2/proto/gateway"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	allowedOriginsFlag = flag.String(
		"gateway-allowed-origins",
		"*",
		"comma-separated, allowed origins for gateway cors",
	)
	grpcServerAddressFlag = flag.String(
		"grpc-server-address",
		"localhost:4500",
		"host:port address for the grpc server",
	)
	grpcGatewayAddressFlag = flag.String(
		"grpc-gateway-address",
		"localhost:5000",
		"host:port address for the grpc-JSON gateway server",
	)
	_ = pb.EventsServer(&server{})
)

type server struct {
	pb.UnimplementedEventsServer
}

func (s *server) StreamEvents(
	req *pb.EventRequest, stream pb.Events_StreamEventsServer,
) error {
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			data, err := anypb.New(req)
			if err != nil {
				return err
			}
			if err := stream.Send(&gwpb.EventSource{
				Event: "attestation",
				Data:  data,
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return errors.New("context canceled")
		}
	}
}

func main() {
	flag.Parse()
	grpcGatewayAddress := *grpcGatewayAddressFlag
	grpcServerAddress := *grpcServerAddressFlag
	lis, err := net.Listen("tcp", grpcServerAddress)
	if err != nil {
		log.Fatalf("Could not listen to port in Start() %s: %v", grpcServerAddress, err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventsServer(grpcServer, &server{})

	go func() {
		log.Printf("gRPC server listening on port: %s", grpcServerAddress)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Could not serve gRPC: %v", err)
		}
	}()

	gwmux := gwruntime.NewServeMux(
		gwruntime.WithMarshalerOption(
			gwruntime.MIMEWildcard, &gwruntime.JSONPb{},
		),
	)
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	ctx := context.Background()
	if err := pb.RegisterEventsHandlerFromEndpoint(ctx, gwmux, grpcServerAddress, dialOpts); err != nil {
		log.Fatalf("Could not register API handler with grpc endpoint: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", corsMiddleware(gwmux))
	gatewayServer := &http.Server{
		Addr:    grpcGatewayAddress,
		Handler: mux,
	}

	go func() {
		log.Printf("Starting gRPC gateway: %s", grpcGatewayAddress)
		if err := gatewayServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed to listen and serve: %v", err)
		}
	}()

	stop := make(chan struct{})
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Println("Got interrupt, shutting down...")
		grpcServer.GracefulStop()
		stop <- struct{}{}
	}()

	// Wait for stop channel to be closed.
	<-stop
}

func corsMiddleware(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodPost, http.MethodGet, http.MethodOptions},
		AllowCredentials: true,
		MaxAge:           600,
		AllowedHeaders:   []string{"*"},
	})
	return c.Handler(h)
}
