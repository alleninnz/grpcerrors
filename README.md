# grpcerrors

Type-safe gRPC error handling: runtime library + codegen plugin.

## Install

```bash
go get github.com/JasperLabs/grpcerrors
go install github.com/JasperLabs/grpcerrors/cmd/protoc-gen-go-errors@latest
```

## Plugin usage

Define errors in proto with `(errors.grpc_code)` annotations:

```protobuf
import "errors/errors.proto";

enum ErrorReason {
  ERROR_REASON_INVALID = 0;
  FUND_NOT_FOUND       = 1 [(errors.grpc_code) = NOT_FOUND];
  HOLDINGS_EXIST       = 2 [(errors.grpc_code) = FAILED_PRECONDITION];
}
```

Add to `buf.gen.yaml`:

```yaml
- local: protoc-gen-go-errors
  out: .
  opt: paths=source_relative
```

Consumer repos need `errors.proto` available at compile time. Add this repo's `proto/` directory as a protoc include path:

```bash
protoc -I path/to/grpcerrors/proto your_errors.proto
```

Generated output:

```go
func ErrorFundNotFound(msg ...string) *grpcerrors.Error  // create
func IsFundNotFound(err error) bool                      // check
```

## Runtime API

| Function | Description |
|---|---|
| `New(code, reason, domain, msg)` | Create error |
| `FromError(err)` | Extract from gRPC status |
| `MatchReasonDomain(err, reason, domain)` | Check match |
| `Reason(err)` / `Code(err)` | Accessors |
| `(*Error).WithMetadata(md)` | Attach metadata |
| `(*Error).GRPCStatus()` | Convert to gRPC status with ErrorInfo |
| `UnaryClientInterceptor()` | Unary client interceptor |
| `StreamClientInterceptor()` | Stream client interceptor |

## Development

```bash
make build    # Build runtime + plugin
make test     # Run all tests
make protoc   # Regenerate errors.pb.go
make install  # Install plugin to $GOPATH/bin
```

## Structure

```
├── errors.go, interceptor.go        # Runtime
├── errors.pb.go                     # Extension registration (generated)
├── proto/errors/errors.proto        # Extension definition (consumer include root)
├── cmd/protoc-gen-go-errors/        # Codegen plugin
│   └── testdata/                    # Golden test fixtures
└── go.mod                           # Single module
```
