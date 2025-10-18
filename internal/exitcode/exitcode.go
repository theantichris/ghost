package exitcode

import "errors"

// ExitCode represents process exit status codes following the sysexits.h convention.
type ExitCode int

const (
	// Ok indicates successful execution.
	Ok ExitCode = 0

	// ExDefault indicates a generic/unspecified error.
	ExDefault ExitCode = 1

	// ExUsage indicates incorrect command line arguments or CLI misuse.
	ExUsage ExitCode = 64

	// ExDataErr indicates invalid input data or content.
	ExDataErr ExitCode = 65

	// ExNoInput indicates a required input file or stream is missing.
	ExNoInput ExitCode = 66

	// ExNoUser indicates a specified user was not found.
	ExNoUser ExitCode = 67

	// ExNoHost indicates a specified host was not found.
	ExNoHost ExitCode = 68

	// ExUnavailable indicates a required service or resource is unavailable.
	ExUnavailable ExitCode = 69

	// ExSoftware indicates an internal software error or bug.
	ExSoftware ExitCode = 70

	// ExOsErr indicates an operating system error such as fork or pipe failure.
	ExOsErr ExitCode = 71

	// ExOsFile indicates a system file is missing or has incorrect permissions.
	ExOsFile ExitCode = 72

	// ExCantCreat indicates the application cannot create the requested output.
	ExCantCreat ExitCode = 73

	// ExIOErr indicates an input/output operation failed.
	ExIOErr ExitCode = 74

	// ExTempFail indicates a temporary failure where retry may succeed.
	ExTempFail ExitCode = 75

	// ExProtocol indicates a protocol violation occurred.
	ExProtocol ExitCode = 76

	// ExNoPerm indicates insufficient permissions for the requested operation.
	ExNoPerm ExitCode = 77

	// ExConfig indicates invalid, missing, or inaccessible configuration.
	ExConfig ExitCode = 78
)

// Error wraps an error with an associated exit code for propagation through the call stack.
type Error struct {
	Err  error
	Code ExitCode
}

// New creates a new Error that wraps the given error with an exit code.
func New(err error, exitCode ExitCode) *Error {
	exitErr := Error{
		Err:  err,
		Code: exitCode,
	}

	return &exitErr
}

// Error implements the error interface by returning the wrapped error's message.
func (err *Error) Error() string {
	return err.Err.Error()
}

// Unwrap returns the wrapped error for use with errors.Is and errors.As.
func (err *Error) Unwrap() error {
	return err.Err
}

// ExitCode returns the exit code associated with this error.
func (err *Error) ExitCode() ExitCode {
	return err.Code
}

// GetExitCode extracts the exit code from an error chain, returning ExDefault if no exit code is found.
func GetExitCode(err error) ExitCode {
	var exitErr *Error
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	return ExDefault
}
