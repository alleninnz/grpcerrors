package grpcerrors

import (
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SupportPackageIsVersion1 is a compile-time guard ensuring generated code
// and the runtime are compatible.
const SupportPackageIsVersion1 = true

// Error is a gRPC-aware error carrying a reason, domain, and optional metadata.
// It implements error, GRPCStatus(), and Is().
type Error struct {
	GRPCCode codes.Code
	Reason   string
	Domain   string
	Message  string
	Metadata map[string]string
}

// New creates an *Error. If msg is empty, the reason is used as the message.
func New(code codes.Code, reason, domain, msg string) *Error {
	if msg == "" {
		msg = reason
	}
	return &Error{
		GRPCCode: code,
		Reason:   reason,
		Domain:   domain,
		Message:  msg,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// GRPCStatus converts this error to a gRPC status with an ErrorInfo detail.
func (e *Error) GRPCStatus() *status.Status {
	if e == nil {
		return status.New(codes.Unknown, "")
	}
	s := status.New(e.GRPCCode, e.Message)
	s, err := s.WithDetails(&errdetails.ErrorInfo{
		Reason:   e.Reason,
		Domain:   e.Domain,
		Metadata: e.Metadata,
	})
	if err != nil {
		// WithDetails only fails on proto marshal errors; fall back to plain status.
		return status.New(e.GRPCCode, e.Message)
	}
	return s
}

// Is supports errors.Is matching by Reason and Domain.
func (e *Error) Is(target error) bool {
	if e == nil {
		return false
	}
	var t *Error
	if !errors.As(target, &t) || t == nil {
		return false
	}
	return e.Reason == t.Reason && e.Domain == t.Domain
}

// WithMetadata returns a copy of the Error with the given metadata.
// The original Error is not modified, making it safe to reuse base errors.
func (e *Error) WithMetadata(md map[string]string) *Error {
	if e == nil {
		return nil
	}
	clone := *e
	clone.Metadata = make(map[string]string, len(md))
	for k, v := range md {
		clone.Metadata[k] = v
	}
	return &clone
}

// FromError extracts an *Error from a gRPC status error by looking for an
// ErrorInfo detail. Returns nil if the error is nil or has no ErrorInfo.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	s, ok := status.FromError(err)
	if !ok {
		return nil
	}
	for _, d := range s.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			return &Error{
				GRPCCode: s.Code(),
				Reason:   info.Reason,
				Domain:   info.Domain,
				Message:  s.Message(),
				Metadata: info.Metadata,
			}
		}
	}
	return nil
}

// MatchReasonDomain returns true if err is an *Error with the given reason and domain.
// Works with both *Error values and raw gRPC status errors containing ErrorInfo.
func MatchReasonDomain(err error, reason, domain string) bool {
	e := FromError(err)
	if e == nil {
		return false
	}
	return e.Reason == reason && e.Domain == domain
}

// Reason returns the reason string if err is an *Error, or "" otherwise.
// Works with both *Error values and raw gRPC status errors containing ErrorInfo.
func Reason(err error) string {
	e := FromError(err)
	if e == nil {
		return ""
	}
	return e.Reason
}

// Code returns the gRPC code if err is an *Error, or codes.Unknown otherwise.
// Works with both *Error values and raw gRPC status errors containing ErrorInfo.
func Code(err error) codes.Code {
	e := FromError(err)
	if e == nil {
		return codes.Unknown
	}
	return e.GRPCCode
}
