package grpcerrors

import (
	"errors"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNew(t *testing.T) {
	err := New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "not found")
	if err.GRPCCode != codes.NotFound {
		t.Errorf("GRPCCode = %v, want NotFound", err.GRPCCode)
	}
	if err.Reason != "FUND_NOT_FOUND" {
		t.Errorf("Reason = %q, want FUND_NOT_FOUND", err.Reason)
	}
	if err.Domain != "jasper.admin.investment.v0" {
		t.Errorf("Domain = %q", err.Domain)
	}
	if err.Message != "not found" {
		t.Errorf("Message = %q, want %q", err.Message, "not found")
	}
}

func TestNew_EmptyMessageDefaultsToReason(t *testing.T) {
	err := New(codes.NotFound, "FUND_NOT_FOUND", "d", "")
	if err.Message != "FUND_NOT_FOUND" {
		t.Errorf("Message = %q, want %q (should default to reason)", err.Message, "FUND_NOT_FOUND")
	}
}

func TestError_Error(t *testing.T) {
	err := New(codes.NotFound, "R", "d", "fund 42")
	if err.Error() != "fund 42" {
		t.Errorf("Error() = %q, want %q", err.Error(), "fund 42")
	}
}

func TestError_GRPCStatus(t *testing.T) {
	err := New(codes.FailedPrecondition, "HOLDINGS_EXIST", "jasper.admin.investment.v0", "fund 42").
		WithMetadata(map[string]string{"fund_id": "42"})

	s := err.GRPCStatus()
	if s.Code() != codes.FailedPrecondition {
		t.Errorf("status code = %v, want FailedPrecondition", s.Code())
	}
	if s.Message() != "fund 42" {
		t.Errorf("status message = %q, want %q", s.Message(), "fund 42")
	}

	var found *errdetails.ErrorInfo
	for _, d := range s.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			found = info
			break
		}
	}
	if found == nil {
		t.Fatal("no ErrorInfo detail found")
	}
	if found.Reason != "HOLDINGS_EXIST" {
		t.Errorf("ErrorInfo.Reason = %q, want HOLDINGS_EXIST", found.Reason)
	}
	if found.Domain != "jasper.admin.investment.v0" {
		t.Errorf("ErrorInfo.Domain = %q", found.Domain)
	}
	if found.Metadata["fund_id"] != "42" {
		t.Errorf("ErrorInfo.Metadata = %v", found.Metadata)
	}
}

func TestFromError_RoundTrip(t *testing.T) {
	original := New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "fund 42").
		WithMetadata(map[string]string{"id": "42"})

	restored := FromError(original.GRPCStatus().Err())
	if restored == nil {
		t.Fatal("FromError returned nil")
	}
	if restored.GRPCCode != codes.NotFound {
		t.Errorf("GRPCCode = %v, want NotFound", restored.GRPCCode)
	}
	if restored.Reason != "FUND_NOT_FOUND" {
		t.Errorf("Reason = %q", restored.Reason)
	}
	if restored.Domain != "jasper.admin.investment.v0" {
		t.Errorf("Domain = %q", restored.Domain)
	}
	if restored.Message != "fund 42" {
		t.Errorf("Message = %q", restored.Message)
	}
	if restored.Metadata["id"] != "42" {
		t.Errorf("Metadata = %v", restored.Metadata)
	}
}

func TestFromError_PlainStatus(t *testing.T) {
	err := status.Error(codes.NotFound, "plain error")
	if FromError(err) != nil {
		t.Error("FromError should return nil for plain status error")
	}
}

func TestFromError_Nil(t *testing.T) {
	if FromError(nil) != nil {
		t.Error("FromError(nil) should return nil")
	}
}

func TestMatchReasonDomain(t *testing.T) {
	err := New(codes.NotFound, "NOT_FOUND", "service.a", "msg")

	if !MatchReasonDomain(err, "NOT_FOUND", "service.a") {
		t.Error("should match same reason+domain")
	}
	if MatchReasonDomain(err, "NOT_FOUND", "service.b") {
		t.Error("should not match same reason with different domain")
	}
	if MatchReasonDomain(err, "OTHER", "service.a") {
		t.Error("should not match different reason")
	}
	if MatchReasonDomain(nil, "X", "d") {
		t.Error("should return false for nil")
	}
}

func TestMatchReasonDomain_RawGRPCStatus(t *testing.T) {
	original := New(codes.NotFound, "NOT_FOUND", "service.a", "msg")
	rawStatusErr := original.GRPCStatus().Err()

	if !MatchReasonDomain(rawStatusErr, "NOT_FOUND", "service.a") {
		t.Error("should match raw gRPC status with correct reason+domain")
	}
	if MatchReasonDomain(rawStatusErr, "NOT_FOUND", "service.b") {
		t.Error("should not match raw gRPC status with wrong domain")
	}
}

func TestReason_RawGRPCStatus(t *testing.T) {
	original := New(codes.NotFound, "FUND_NOT_FOUND", "d", "msg")
	rawStatusErr := original.GRPCStatus().Err()

	if got := Reason(rawStatusErr); got != "FUND_NOT_FOUND" {
		t.Errorf("Reason(raw status) = %q, want FUND_NOT_FOUND", got)
	}
}

func TestCode_RawGRPCStatus(t *testing.T) {
	original := New(codes.NotFound, "FUND_NOT_FOUND", "d", "msg")
	rawStatusErr := original.GRPCStatus().Err()

	if got := Code(rawStatusErr); got != codes.NotFound {
		t.Errorf("Code(raw status) = %v, want NotFound", got)
	}
}

func TestReason(t *testing.T) {
	err := New(codes.NotFound, "FUND_NOT_FOUND", "d", "msg")
	if got := Reason(err); got != "FUND_NOT_FOUND" {
		t.Errorf("Reason = %q, want FUND_NOT_FOUND", got)
	}
	if got := Reason(errors.New("plain")); got != "" {
		t.Errorf("Reason(plain) = %q, want empty", got)
	}
}

func TestCode(t *testing.T) {
	err := New(codes.NotFound, "R", "d", "msg")
	if got := Code(err); got != codes.NotFound {
		t.Errorf("Code = %v, want NotFound", got)
	}
	if got := Code(errors.New("plain")); got != codes.Unknown {
		t.Errorf("Code(plain) = %v, want Unknown", got)
	}
}

func TestWithMetadata(t *testing.T) {
	err := New(codes.NotFound, "R", "d", "msg")
	ret := err.WithMetadata(map[string]string{"k": "v"})
	if ret == err {
		t.Error("WithMetadata should return a new *Error, not mutate the original")
	}
	if ret.Metadata["k"] != "v" {
		t.Errorf("returned Metadata = %v", ret.Metadata)
	}
	if err.Metadata != nil {
		t.Errorf("original Metadata should be nil, got %v", err.Metadata)
	}
}

func TestWithMetadata_DeepCopy(t *testing.T) {
	md := map[string]string{"k": "v"}
	err := New(codes.NotFound, "R", "d", "msg").WithMetadata(md)
	md["k"] = "mutated"
	if err.Metadata["k"] != "v" {
		t.Error("WithMetadata should deep-copy the map; caller mutation leaked in")
	}
}

func TestIs(t *testing.T) {
	a := New(codes.NotFound, "X", "d", "m1")
	b := New(codes.Internal, "X", "d", "m2")
	c := New(codes.NotFound, "Y", "d", "m1")
	d := New(codes.NotFound, "X", "other-domain", "m1")

	if !errors.Is(a, b) {
		t.Error("errors.Is should match same reason+domain even with different codes")
	}
	if errors.Is(a, c) {
		t.Error("errors.Is should not match different reasons")
	}
	if errors.Is(a, d) {
		t.Error("errors.Is should not match same reason with different domain")
	}
	if errors.Is(a, errors.New("plain")) {
		t.Error("errors.Is should not match non-Error")
	}

	// Typed-nil target regression: must not panic.
	var nilTarget *Error
	if errors.Is(a, nilTarget) {
		t.Error("errors.Is should not match typed-nil target")
	}

	// Typed-nil receiver regression: must not panic.
	var nilReceiver *Error
	if errors.Is(nilReceiver, a) {
		t.Error("errors.Is should not match typed-nil receiver")
	}
}

func TestNilReceiver_AllMethods(t *testing.T) {
	var e *Error

	if e.Error() != "" {
		t.Error("nil.Error() should return empty string")
	}

	s := e.GRPCStatus()
	if s.Code() != codes.Unknown {
		t.Errorf("nil.GRPCStatus().Code() = %v, want Unknown", s.Code())
	}

	if e.WithMetadata(map[string]string{"k": "v"}) != nil {
		t.Error("nil.WithMetadata() should return nil")
	}
}
