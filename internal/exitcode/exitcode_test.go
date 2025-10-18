package exitcode

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")
	expectedCode := ExIOErr

	exitErr := New(expectedErr, expectedCode)

	actualErr := exitErr.Err

	if !errors.Is(actualErr, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, actualErr)
	}

	actualCode := exitErr.Code

	if actualCode != expectedCode {
		t.Errorf("expected exit code %d, got %d", expectedCode, actualCode)
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")

	exitErr := Error{
		Err: expectedErr,
	}

	if exitErr.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr.Error(), exitErr.Error())
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("test error")

	exitErr := Error{
		Err: expectedErr,
	}

	actualErr := exitErr.Unwrap()

	if !errors.Is(actualErr, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, actualErr)
	}
}

func TestExitCode(t *testing.T) {
	t.Parallel()

	expectedCode := ExIOErr

	exitErr := Error{
		Code: expectedCode,
	}

	actualCode := exitErr.ExitCode()

	if actualCode != expectedCode {
		t.Errorf("expected code %d, got %d", expectedCode, actualCode)
	}
}

func TestGetExitCode(t *testing.T) {
	t.Run("returns exit code for known error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("test error")
		expectedCode := ExIOErr

		exitErr := New(err, expectedCode)

		actualCode := GetExitCode(exitErr)

		if actualCode != expectedCode {
			t.Errorf("expected exit code %d, got %d", expectedCode, actualCode)
		}
	})

	t.Run("returns exit code 1 for unknown error", func(t *testing.T) {
		t.Parallel()

		err := errors.New("test error")

		actualCode := GetExitCode(err)

		if actualCode != ExDefault {
			t.Errorf("expected error code %d, got %d", ExDefault, actualCode)
		}
	})
}
