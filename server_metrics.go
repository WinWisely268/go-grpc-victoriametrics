package grpc_victoriametrics

import (
	"context"
	"github.com/VictoriaMetrics/metrics"
	"github.com/grpc-ecosystem/go-grpc-prometheus/packages/grpcstatus"
	"google.golang.org/grpc"
)

const (
	METRICS_STARTED_EXP_FMT     = `grpc_server_started_total{grpc_type="%s",grpc_service="%s",grpc_method="%s"}`
	METRICS_HANDLED_EXP_FMT     = `grpc_server_started_total{grpc_type="%s",grpc_service="%s",grpc_method="%s",grpc_code="%s"}`
	METRICS_STREAM_RECV_EXP_FMT = `grpc_server_msg_received_total{grpc_type="%s",grpc_service="%s",grpc_method="%s"}`
	METRICS_STREAM_SENT_EXP_FMT = `grpc_server_msg_sent_total{grpc_type="%s",grpc_service="%s",grpc_method="%s"}`
	METRICS_HANDLE_TIME_EXP_FMT = `grpc_server_handling_seconds{grpc_type="%s",grpc_service="%s",grpc_method="%s"}`
)

// ServerMetrics represents a collection of metrics to be registered
// for a gRPC server.
type ServerMetrics struct {
	enableHistogram bool
}

// NewServerMetrics returns a ServerMetrics object.
func NewServerMetrics(enableHistogram bool) *ServerMetrics {
	return &ServerMetrics{
		enableHistogram,
	}
}
