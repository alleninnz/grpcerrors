# grpcerrors

Runtime library for type-safe gRPC error handling in Go. Pairs with [protoc-gen-go-errors](https://github.com/JasperLabs/protoc-gen-go-errors) to provide generated error constructors and checkers.

## Install

```bash
go get github.com/JasperLabs/grpcerrors
```

## Usage

### Creating errors

```go
// With message
err := grpcerrors.New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "fund not found")

// Empty message defaults to reason string
err := grpcerrors.New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "")
// err.Message == "FUND_NOT_FOUND"

// With metadata
err := grpcerrors.New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "fund 42").
    WithMetadata(map[string]string{"fund_id": "42"})
```

### Checking errors

```go
// Match by reason + domain
if grpcerrors.MatchReasonDomain(err, "FUND_NOT_FOUND", "jasper.admin.investment.v0") { ... }

// Accessors
reason := grpcerrors.Reason(err)   // "FUND_NOT_FOUND" or ""
code := grpcerrors.Code(err)       // codes.NotFound or codes.Unknown
```

### Wire format

`*Error` implements `GRPCStatus()` — gRPC-Go automatically converts it to a status with:
- gRPC status code and message
- `google.rpc.ErrorInfo` detail carrying `reason`, `domain`, and `metadata`

No server-side interceptor needed.

### Client interceptor

Reconstitutes `*Error` from gRPC status details on the receiving side:

```go
conn, _ := grpc.Dial(addr,
    grpc.WithChainUnaryInterceptor(grpcerrors.UnaryClientInterceptor()),
    grpc.WithChainStreamInterceptor(grpcerrors.StreamClientInterceptor()),
)
```

The interceptors wrap all error-returning stream methods (`RecvMsg`, `SendMsg`, `Header`, `CloseSend`).

`MatchReasonDomain` and other accessors also work on raw gRPC status errors without the interceptor — they extract `ErrorInfo` on the fly.

## API

| Function | Description |
|---|---|
| `New(code, reason, domain, msg)` | Create error (empty msg defaults to reason) |
| `FromError(err)` | Extract `*Error` from gRPC status (nil if no ErrorInfo) |
| `MatchReasonDomain(err, reason, domain)` | Check reason + domain match |
| `Reason(err)` | Get reason string |
| `Code(err)` | Get gRPC code |
| `(*Error).WithMetadata(md)` | Return copy with metadata (immutable) |
| `(*Error).GRPCStatus()` | Convert to gRPC status with ErrorInfo detail |
| `UnaryClientInterceptor()` | Unary client interceptor |
| `StreamClientInterceptor()` | Stream client interceptor |
