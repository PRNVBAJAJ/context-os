package shared_test

import (
	"errors"
	"testing"

	"github.com/PRNVBAJAJ/context-os/internal/shared"
)

func TestNewError_Format(t *testing.T) {
	tests := []struct {
		name    string
		code    shared.Code
		message string
		want    string
	}{
		{
			name:    "not found",
			code:    shared.CodeNotFound,
			message: "workflow not found",
			want:    "[NOT_FOUND] workflow not found",
		},
		{
			name:    "invalid input",
			code:    shared.CodeInvalidInput,
			message: "name is required",
			want:    "[INVALID_INPUT] name is required",
		},
		{
			name:    "conflict",
			code:    shared.CodeConflict,
			message: "project already initialized",
			want:    "[CONFLICT] project already initialized",
		},
		{
			name:    "internal",
			code:    shared.CodeInternal,
			message: "unexpected failure",
			want:    "[INTERNAL] unexpected failure",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := shared.NewError(tc.code, tc.message)

			if err.Error() != tc.want {
				t.Errorf("Error() = %q, want %q", err.Error(), tc.want)
			}
			if err.Code != tc.code {
				t.Errorf("Code = %q, want %q", err.Code, tc.code)
			}
			if err.Cause != nil {
				t.Error("Cause should be nil for NewError")
			}
		})
	}
}

func TestWrap_FormatIncludesCause(t *testing.T) {
	cause := errors.New("disk full")
	err := shared.Wrap(shared.CodeInternal, "storage write failed", cause)

	want := "[INTERNAL] storage write failed: disk full"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestWrap_ErrorsIs(t *testing.T) {
	cause := errors.New("connection refused")
	err := shared.Wrap(shared.CodeInternal, "storage unavailable", cause)

	if !errors.Is(err, cause) {
		t.Error("errors.Is should traverse the cause chain")
	}
}

func TestWrap_ErrorsAs(t *testing.T) {
	inner := shared.NewError(shared.CodeNotFound, "project not found")
	outer := shared.Wrap(shared.CodeInternal, "init failed", inner)

	var target *shared.Error
	if !errors.As(outer, &target) {
		t.Fatal("errors.As should unwrap to *shared.Error")
	}
	// The first match is the outer error itself.
	if target.Code != shared.CodeInternal {
		t.Errorf("Code = %q, want %q", target.Code, shared.CodeInternal)
	}
}

func TestError_ImplementsError(t *testing.T) {
	// Compile-time assertion: *Error satisfies the error interface.
	var _ error = (*shared.Error)(nil)
}
