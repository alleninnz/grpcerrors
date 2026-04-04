package grpcerrors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryClientInterceptor returns a gRPC unary client interceptor that
// reconstitutes *Error from gRPC status details containing ErrorInfo.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}
		if e := FromError(err); e != nil {
			return e
		}
		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor that
// reconstitutes *Error from gRPC status details on stream creation,
// RecvMsg, SendMsg, Header, and CloseSend.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			if e := FromError(err); e != nil {
				return s, e
			}
			return s, err
		}
		return &wrappedStream{ClientStream: s}, nil
	}
}

type wrappedStream struct {
	grpc.ClientStream
}

// reconstitute converts a raw gRPC status error to *Error if it contains ErrorInfo.
func reconstitute(err error) error {
	if err == nil {
		return nil
	}
	if e := FromError(err); e != nil {
		return e
	}
	return err
}

func (w *wrappedStream) RecvMsg(m any) error {
	return reconstitute(w.ClientStream.RecvMsg(m))
}

func (w *wrappedStream) SendMsg(m any) error {
	return reconstitute(w.ClientStream.SendMsg(m))
}

func (w *wrappedStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	return md, reconstitute(err)
}

func (w *wrappedStream) CloseSend() error {
	return reconstitute(w.ClientStream.CloseSend())
}
