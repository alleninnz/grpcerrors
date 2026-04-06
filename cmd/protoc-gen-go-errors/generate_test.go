package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func requireProtoc(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skipf(
			"protoc not found in PATH: %v (install with: apt-get install protobuf-compiler)",
			err,
		)
	}
}

func TestGoldenFile(t *testing.T) {
	requireProtoc(t)
	// Build the plugin binary.
	pluginBin := filepath.Join(t.TempDir(), "protoc-gen-go-errors")
	build := exec.Command("go", "build", "-o", pluginBin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, out)
	}

	// Run protoc with the plugin against testdata/test.proto.
	outDir := t.TempDir()
	protoDir := filepath.Join("..", "..", "proto")
	testdataDir := filepath.Join("testdata")

	protoc := exec.Command("protoc",
		"--plugin=protoc-gen-go-errors="+pluginBin,
		"--go-errors_out="+outDir,
		"--go-errors_opt=paths=source_relative",
		"-I", protoDir,
		"-I", testdataDir,
		filepath.Join(testdataDir, "test.proto"),
	)
	if out, err := protoc.CombinedOutput(); err != nil {
		t.Fatalf("protoc: %v\n%s", err, out)
	}

	// Read generated output.
	gotPath := filepath.Join(outDir, "test_errors.pb.go")
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}

	goldenPath := filepath.Join(testdataDir, "test_errors.pb.go")

	if *update {
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatalf("update golden file: %v", err)
		}
		t.Log("updated golden file")
		return
	}

	// Compare with golden file.
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v (run with -update to create)", err)
	}

	if string(got) != string(want) {
		diff, _ := exec.Command("diff", "-u", goldenPath, gotPath).CombinedOutput()
		t.Errorf("generated output differs from golden file:\n%s", diff)
	}
}

func TestNoOutputForUnannotatedProto(t *testing.T) {
	requireProtoc(t)
	// Build the plugin binary.
	pluginBin := filepath.Join(t.TempDir(), "protoc-gen-go-errors")
	build := exec.Command("go", "build", "-o", pluginBin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, out)
	}

	outDir := t.TempDir()
	testdataDir := filepath.Join("testdata")

	protoc := exec.Command("protoc",
		"--plugin=protoc-gen-go-errors="+pluginBin,
		"--go-errors_out="+outDir,
		"--go-errors_opt=paths=source_relative",
		"-I", testdataDir,
		filepath.Join(testdataDir, "no_errors.proto"),
	)
	if out, err := protoc.CombinedOutput(); err != nil {
		t.Fatalf("protoc: %v\n%s", err, out)
	}

	// Verify no output file was generated.
	outFile := filepath.Join(outDir, "no_errors_errors.pb.go")
	if _, err := os.Stat(outFile); err == nil {
		t.Errorf("expected no output file for unannotated proto, but %s exists", outFile)
	}
}

func TestNestedEnumGeneratesHelpers(t *testing.T) {
	requireProtoc(t)
	pluginBin := filepath.Join(t.TempDir(), "protoc-gen-go-errors")
	build := exec.Command("go", "build", "-o", pluginBin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, out)
	}

	outDir := t.TempDir()
	protoDir := filepath.Join("..", "..", "proto")
	testdataDir := filepath.Join("testdata")

	protoc := exec.Command("protoc",
		"--plugin=protoc-gen-go-errors="+pluginBin,
		"--go-errors_out="+outDir,
		"--go-errors_opt=paths=source_relative",
		"-I", protoDir,
		"-I", testdataDir,
		filepath.Join(testdataDir, "nested.proto"),
	)
	if out, err := protoc.CombinedOutput(); err != nil {
		t.Fatalf("protoc: %v\n%s", err, out)
	}

	outFile := filepath.Join(outDir, "nested_errors.pb.go")
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file for nested enum, but got: %v", err)
	}

	content := string(got)
	for _, want := range []string{"ErrorFundHasHoldings", "IsFundHasHoldings", "ErrorFundHasOffers", "IsFundHasOffers"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in generated output, not found", want)
		}
	}
}

func TestDuplicateHelperNameFails(t *testing.T) {
	requireProtoc(t)
	pluginBin := filepath.Join(t.TempDir(), "protoc-gen-go-errors")
	build := exec.Command("go", "build", "-o", pluginBin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, out)
	}

	outDir := t.TempDir()
	protoDir := filepath.Join("..", "..", "proto")
	testdataDir := filepath.Join("testdata")

	// duplicate.proto has two nested enums with NOT_FOUND — both produce
	// ErrorNotFound/IsNotFound. Plugin should fail with a clear error.
	protoc := exec.Command("protoc",
		"--plugin=protoc-gen-go-errors="+pluginBin,
		"--go-errors_out="+outDir,
		"--go-errors_opt=paths=source_relative",
		"-I", protoDir,
		"-I", testdataDir,
		filepath.Join(testdataDir, "duplicate.proto"),
	)
	out, err := protoc.CombinedOutput()
	if err == nil {
		t.Fatal("expected plugin to fail on duplicate helper names, but it succeeded")
	}
	if !strings.Contains(string(out), "duplicate helper name") {
		t.Errorf("expected 'duplicate helper name' in error output, got:\n%s", out)
	}
}

func TestInvalidGRPCCodeRejectedByProtoc(t *testing.T) {
	requireProtoc(t)
	protoDir := filepath.Join("..", "..", "proto")
	testdataDir := filepath.Join("testdata")

	// protoc should reject invalid_code.proto because NONE_EXISTED_CODE
	// is not a valid google.rpc.Code enum value.
	protoc := exec.Command("protoc",
		"--descriptor_set_out="+os.DevNull,
		"-I", protoDir,
		"-I", testdataDir,
		filepath.Join(testdataDir, "invalid_code.proto"),
	)
	out, err := protoc.CombinedOutput()
	if err == nil {
		t.Fatal("expected protoc to reject invalid gRPC code, but it succeeded")
	}
	output := string(out)
	if !strings.Contains(output, "NONE_EXISTED_CODE") {
		t.Errorf("expected error to mention NONE_EXISTED_CODE, got:\n%s", output)
	}
}
