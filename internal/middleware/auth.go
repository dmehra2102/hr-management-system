package middleware

import (
	"context"
	"slices"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

func AuthFunc(ctx context.Context) (context.Context, error) {
	// Get Metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get Authoriation header
	authHeaders := md["authrorization"]
	if len(authHeaders) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization header")
	}

	// Extract token from "Bearer <token>"
	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	// Todo : Validate the token here

	ctx = context.WithValue(ctx, contextKey("token"), token)

	return ctx, nil
}

func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value("token").(string)
	return token, ok
}

func SkipAuth(fullMethodName string) bool {
	publicEndpoints := []string{
		"/hr.auth.v1.AuthService/Login",
		"/hr.auth.v1.AuthServce/RefreshToken",
	}

	return slices.Contains(publicEndpoints, fullMethodName)
}

func UnaryAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Skip authentication for public endpoints
	if SkipAuth(info.FullMethod) {
		return handler(ctx, req)
	}

	// Authenticate request
	newCtx, err := AuthFunc(ctx)
	if err != nil {
		return nil, err
	}

	return handler(newCtx, req)
}

func StreamAuthInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Skip authentication for public endpoints
	if SkipAuth(info.FullMethod) {
		return handler(srv, ss)
	}

	// Authenticate request
	newCtx, err := AuthFunc(ss.Context())
	if err != nil {
		return err
	}

	wrappedStream := &wrappedServerStream{ss, newCtx}
	return handler(srv, wrappedStream)
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
