package grpcerrors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptor_ReconstitutesError(t *testing.T) {
	original := Newf(codes.NotFound, "FUND_NOT_FOUND", "d", "msg")
	interceptor := UnaryClientInterceptor()

	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return original.GRPCStatus().Err()
		},
	)

	if !MatchReasonDomain(err, "FUND_NOT_FOUND", "d") {
		t.Errorf("expected reconstituted *Error with reason FUND_NOT_FOUND, got %v", err)
	}
}

func TestUnaryClientInterceptor_PassesPlainError(t *testing.T) {
	plainErr := status.Error(codes.NotFound, "plain")
	interceptor := UnaryClientInterceptor()

	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return plainErr
		},
	)

	if _, ok := err.(*Error); ok {
		t.Error("expected plain status error, got *Error")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected NotFound, got %v", status.Code(err))
	}
}

func TestUnaryClientInterceptor_PassesNil(t *testing.T) {
	interceptor := UnaryClientInterceptor()

	err := interceptor(context.Background(), "/test", nil, nil, nil,
		func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		},
	)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestStreamClientInterceptor_PassesStreamerError(t *testing.T) {
	original := Newf(codes.Unavailable, "SERVICE_DOWN", "d", "msg")
	interceptor := StreamClientInterceptor()

	_, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, original.GRPCStatus().Err()
		},
	)

	if !MatchReasonDomain(err, "SERVICE_DOWN", "d") {
		t.Errorf("expected reconstituted *Error with reason SERVICE_DOWN, got %v", err)
	}
}

// mockStream implements grpc.ClientStream with configurable errors.
type mockStream struct {
	grpc.ClientStream
	recvErr      error
	sendErr      error
	headerErr    error
	headerMD     metadata.MD
	closeSendErr error
}

func (m *mockStream) RecvMsg(msg any) error {
	return m.recvErr
}

func (m *mockStream) SendMsg(msg any) error {
	return m.sendErr
}

func (m *mockStream) Header() (metadata.MD, error) {
	return m.headerMD, m.headerErr
}

func (m *mockStream) CloseSend() error {
	return m.closeSendErr
}

func TestStreamClientInterceptor_RecvMsgReconstitutes(t *testing.T) {
	original := Newf(codes.FailedPrecondition, "HOLDINGS_EXIST", "d", "msg")
	interceptor := StreamClientInterceptor()

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockStream{recvErr: original.GRPCStatus().Err()}, nil
		},
	)
	if err != nil {
		t.Fatalf("streamer returned error: %v", err)
	}

	recvErr := stream.RecvMsg(nil)
	if !MatchReasonDomain(recvErr, "HOLDINGS_EXIST", "d") {
		t.Errorf("expected reconstituted *Error with reason HOLDINGS_EXIST, got %v", recvErr)
	}
}

func TestStreamClientInterceptor_SendMsgReconstitutes(t *testing.T) {
	original := Newf(codes.FailedPrecondition, "FUND_LOCKED", "d", "msg")
	interceptor := StreamClientInterceptor()

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockStream{sendErr: original.GRPCStatus().Err()}, nil
		},
	)
	if err != nil {
		t.Fatalf("streamer returned error: %v", err)
	}

	sendErr := stream.SendMsg(nil)
	if !MatchReasonDomain(sendErr, "FUND_LOCKED", "d") {
		t.Errorf("expected reconstituted *Error with reason FUND_LOCKED, got %v", sendErr)
	}
}

func TestStreamClientInterceptor_HeaderReconstitutes(t *testing.T) {
	original := Newf(codes.Unauthenticated, "TOKEN_EXPIRED", "d", "msg")
	interceptor := StreamClientInterceptor()

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockStream{headerErr: original.GRPCStatus().Err()}, nil
		},
	)
	if err != nil {
		t.Fatalf("streamer returned error: %v", err)
	}

	_, headerErr := stream.Header()
	if !MatchReasonDomain(headerErr, "TOKEN_EXPIRED", "d") {
		t.Errorf("expected reconstituted *Error with reason TOKEN_EXPIRED, got %v", headerErr)
	}
}

func TestStreamClientInterceptor_CloseSendReconstitutes(t *testing.T) {
	original := Newf(codes.Unavailable, "STREAM_CLOSED", "d", "msg")
	interceptor := StreamClientInterceptor()

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/test",
		func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockStream{closeSendErr: original.GRPCStatus().Err()}, nil
		},
	)
	if err != nil {
		t.Fatalf("streamer returned error: %v", err)
	}

	closeErr := stream.CloseSend()
	if !MatchReasonDomain(closeErr, "STREAM_CLOSED", "d") {
		t.Errorf("expected reconstituted *Error with reason STREAM_CLOSED, got %v", closeErr)
	}
}
