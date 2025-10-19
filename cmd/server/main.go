package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmehra2102/hr-management-system/internal/config"
	"github.com/dmehra2102/hr-management-system/internal/database"
	"github.com/dmehra2102/hr-management-system/internal/department"
	"github.com/dmehra2102/hr-management-system/internal/employee"
	"github.com/dmehra2102/hr-management-system/internal/middleware"
	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	departmentpb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/department"
	employeepb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/employee"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpctags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

type Server struct {
	config     *config.Config
	logger     *logger.Logger
	db         *database.Database
	grpcServer *grpc.Server
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	log.Info("Starting HR Management System", "version", "1.0.0", "env", cfg.AppEnv)

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Failed to close database connection", "error", err)
		}
	}()

	if err := db.Migrate(); err != nil {
		log.Error("Failed to run database migration", "error", err)
		os.Exit(1)
	}

	server := &Server{
		config: cfg,
		logger: log,
		db:     db,
	}

	// Start server
	if err := server.Start(); err != nil {
		log.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func (s *Server) Start() error {

	s.grpcServer = grpc.NewServer(
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			grpctags.StreamServerInterceptor(),
			grpclogrus.StreamServerInterceptor(s.logger.GetLogrusEntry()),
			grpcauth.StreamServerInterceptor(middleware.AuthFunc),
			grpcrecovery.StreamServerInterceptor(),
			middleware.StreamRecoveryInterceptor(s.logger),
		)),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			grpctags.UnaryServerInterceptor(),
			grpclogrus.UnaryServerInterceptor(s.logger.GetLogrusEntry()),
			grpcauth.UnaryServerInterceptor(middleware.AuthFunc),
			grpcrecovery.UnaryServerInterceptor(),
			middleware.RecoveryInterceptor(s.logger),
		)),
	)

	s.registerServices()

	// Enable reflection for development
	if s.config.AppEnv == "development" {
		reflection.Register(s.grpcServer)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.GRPCPort, err)
	}

	s.logger.Info("gRPC server starting", "port", s.config.GRPCPort)

	errChan := make(chan error, 1)
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-signalCh:
		s.logger.Info("Shutdown signal received")
		return s.shutdown()
	}
}

func (s *Server) registerServices() {
	// jwtService := auth.NewJWTService(s.config.JWTSecret, time.Duration(s.config.JWTExpiryHours) * time.Hour)

	employeeRepo := employee.NewRepository(s.db.GetDB())
	departmentRepo := department.NewRepository(s.db.GetDB())

	employeeService := employee.NewService(employeeRepo, s.logger)
	departmentService := department.NewService(departmentRepo, s.logger)

	employeeHandler := employee.NewHandler(employeeService, s.logger)
	departmentHandler := department.NewHandler(departmentService,s.logger)

	employeepb.RegisterEmployeeServiceServer(s.grpcServer, employeeHandler)
	departmentpb.RegisterDepartmentServiceServer(s.grpcServer, departmentHandler)
	
	s.logger.Info("All gRPC services registered successfully")
}

func (s *Server) shutdown() error {
	s.logger.Info("Graxeful shutdown completed")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdownCh := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(shutdownCh)
	}()

	select {
	case <-shutdownCh:
		return nil
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout exceeded, forcing shutdown")
		s.grpcServer.Stop()
		return ctx.Err()
	}
}
