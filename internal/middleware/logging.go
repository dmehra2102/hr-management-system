package middleware

import (
	"context"
	"time"

	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()

		requestID := getRequestID(ctx)

		reqLogger := log.WithFields(map[string]any{
			"request_id": requestID,
			"method":     info.FullMethod,
			"component":  "grpc",
		})

		reqLogger.Info("Request started")

		resp, err = handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.WithFields(map[string]any{
				"duration_ms": duration.Milliseconds(),
				"status_code": st.Code(),
				"error":       err.Error(),
			}).Error("Request failed")
		} else {
			reqLogger.WithFields(map[string]any{
				"duration_ms": duration.Milliseconds(),
				"status_code": codes.OK,
			}).Info("Request completed")
		}

		return resp, err
	}

}

func StreamLoggingInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		requestID := getRequestID(ss.Context())

		reqLogger := log.WithFields(map[string]any{
			"request_id": requestID,
			"method":     info.FullMethod,
			"component":  "grpc-stream",
		})

		reqLogger.Info("Stream started")

		err := handler(srv, ss)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.WithFields(map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"status_code": st.Code(),
				"error":       err.Error(),
			}).Error("Stream failed")
		} else {
			reqLogger.WithFields(map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"status_code": codes.OK,
			}).Info("Stream completed")
		}

		return err
	}
}

func RecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(map[string]any{
					"method": info.FullMethod,
					"panic":  r,
				}).Error("Panic recovered")

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

func StreamRecoveryInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(map[string]interface{}{
					"method": info.FullMethod,
					"panic":  r,
				}).Error("Stream panic recovered")

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}

	return generateRequestID()
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range charset {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
