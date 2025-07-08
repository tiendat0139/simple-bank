package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error()
	}
	logger.Str("protocol", "grpc").
		Str("method", info.FullMethod).
		Dur("duration", duration).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Msg("gRPC request received")

	return result, err
}

type ResponseRecoder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *ResponseRecoder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *ResponseRecoder) Write(b []byte) (int, error) {
	r.body = b
	return r.ResponseWriter.Write(b)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecoder{
			ResponseWriter: res,
			statusCode:     http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.statusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.body)
		}

		logger.Str("protocol", "grpc").
			Str("method", req.Method).
			Str("path", req.RequestURI).
			Int("status_code", rec.statusCode).
			Str("status_text", http.StatusText(rec.statusCode)).
			Dur("duration", duration).
			Msg("gRPC request received")
	})
}
