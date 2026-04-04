# grpcerrors

Runtime library for type-safe gRPC error handling in Go. Pairs with [protoc-gen-go-errors](https://github.com/JasperLabs/protoc-gen-go-errors) to provide generated error constructors and checkers.

## Install

```bash
go get github.com/JasperLabs/grpcerrors
```

## Usage

### Creating errors

```go
// Static message
err := grpcerrors.New(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "fund not found")

// Formatted message
err := grpcerrors.Newf(codes.FailedPrecondition, "HOLDINGS_EXIST", "jasper.admin.investment.v0", "fund %d has %d holdings", fundID, count)

// With metadata
err := grpcerrors.Newf(codes.NotFound, "FUND_NOT_FOUND", "jasper.admin.investment.v0", "fund %d", fundID).
    WithMetadata(map[string]string{"fund_id": strconv.Itoa(fundID)})
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
| `New(code, reason, domain, msg)` | Create error with static message |
| `Newf(code, reason, domain, format, args...)` | Create error with formatted message |
| `FromError(err)` | Extract `*Error` from gRPC status (nil if no ErrorInfo) |
| `MatchReasonDomain(err, reason, domain)` | Check reason + domain match |
| `Reason(err)` | Get reason string |
| `Code(err)` | Get gRPC code |
| `(*Error).WithMetadata(md)` | Return copy with metadata (immutable) |
| `(*Error).GRPCStatus()` | Convert to gRPC status with ErrorInfo detail |
| `UnaryClientInterceptor()` | Unary client interceptor |
| `StreamClientInterceptor()` | Stream client interceptor |
