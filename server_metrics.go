package grpc_victoriametrics

import (
	"context"
	"fmt"
	"github.com/VictoriaMetrics/metrics"
	"github.com/winwisely268/go-grpc-victoriametrics/packages/grpcstatus"
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

// counterHelper is just a helper to create a counter for victoriametrics
func counterHelper(expFormat string, labels ...interface{}) *metrics.Counter {
	return metrics.GetOrCreateCounter(fmt.Sprintf(expFormat, labels...))
}

// histHelper is just a helper to create a histogram for victoriametrics
func histHelper(expFormat string, labels ...interface{}) *metrics.Histogram {
	return metrics.GetOrCreateHistogram(fmt.Sprintf(expFormat, labels...))
}

// counterHelper is just a helper to create a counter for victoriametrics
func (m *ServerMetrics) counterHelper(expFormat string, labels ...interface{}) *metrics.Counter {
	return counterHelper(fmt.Sprintf(expFormat, labels...))
}

// histHelper is just a helper to create a histogram for victoriametrics
func (m *ServerMetrics) histHelper(expFormat string, labels ...interface{}) *metrics.Histogram {
	return histHelper(fmt.Sprintf(expFormat, labels...))
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ServerMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		monitor := newServerReporter(m, Unary, info.FullMethod)
		monitor.ReceivedMessage()
		resp, err := handler(ctx, req)
		st, _ := grpcstatus.FromError(err)
		monitor.Handled(st.Code())
		if err == nil {
			monitor.SentMessage()
		}
		return resp, err
	}
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ServerMetrics) StreamServerInterceptor() func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		monitor := newServerReporter(m, streamRPCType(info), info.FullMethod)
		err := handler(srv, &monitoredServerStream{ss, monitor})
		st, _ := grpcstatus.FromError(err)
		monitor.Handled(st.Code())
		return err
	}
}

// InitializeMetrics initializes all metrics, with their appropriate null
// value, for all gRPC methods registered on a gRPC server. This is useful, to
// ensure that all metrics exist when collecting and querying.
func (m *ServerMetrics) InitializeMetrics(server *grpc.Server) {
	serviceInfo := server.GetServiceInfo()
	for serviceName, info := range serviceInfo {
		for _, mInfo := range info.Methods {
			preRegisterMethod(m, serviceName, &mInfo)
		}
	}
}

func streamRPCType(info *grpc.StreamServerInfo) grpcType {
	if info.IsClientStream && !info.IsServerStream {
		return ClientStream
	} else if !info.IsClientStream && info.IsServerStream {
		return ServerStream
	}
	return BidiStream
}

// monitoredStream wraps grpc.ServerStream allowing each Sent/Recv of message to increment counters.
type monitoredServerStream struct {
	grpc.ServerStream
	monitor *serverReporter
}

func (s *monitoredServerStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.monitor.SentMessage()
	}
	return err
}

func (s *monitoredServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.monitor.ReceivedMessage()
	}
	return err
}

// preRegisterMethod is invoked on Register of a Server, allowing all gRPC services labels to be pre-populated.
func preRegisterMethod(m *ServerMetrics, serviceName string, mInfo *grpc.MethodInfo) {
	methodName := mInfo.Name
	methodType := string(typeFromMethodInfo(mInfo))
	// These are just references (no increments), as just referencing will create the labels but not set values.
	counterHelper(METRICS_STARTED_EXP_FMT, methodType, serviceName, methodName)
	counterHelper(METRICS_STREAM_RECV_EXP_FMT, methodType, serviceName, methodName)
	counterHelper(METRICS_STREAM_SENT_EXP_FMT, methodType, serviceName, methodName)
	if m.enableHistogram {
		histHelper(METRICS_HANDLE_TIME_EXP_FMT, methodType, serviceName, methodName)
	}
	for _, code := range allCodes {
		counterHelper(METRICS_HANDLED_EXP_FMT, methodType, serviceName, methodName, code.String())
	}
}
