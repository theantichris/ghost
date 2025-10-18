package exitcode

// ExitCode represents exit codes.
type ExitCode int

const (
	// Ok/Success
	Ok ExitCode = 0

	// Bag flags/arguments, misuse of CLI
	ExUsage ExitCode = 64

	// Invalid input data/content
	ExDataErr ExitCode = 65

	// Missing input file/stream
	ExNoInput ExitCode = 66

	// User not found
	ExNoUser ExitCode = 67

	// Host not found
	ExNoHost ExitCode = 68

	// Service/resource unavailable
	ExUnavailable ExitCode = 69

	// Internal error/bug
	ExSoftware ExitCode = 70

	// Operating system failure: fork, pipe, etc.
	ExOsErr ExitCode = 71

	// Operating system file missing/permission
	ExOsFile ExitCode = 72

	// Can't create output
	ExCantCreat ExitCode = 73

	// Input/output error
	ExIOErr ExitCode = 74

	// Temporary failure, retry may succeed
	ExTempFail ExitCode = 75

	// Protocol violation
	ExProtocol ExitCode = 76

	// Permission denied
	ExNoPerm ExitCode = 77

	// Bad/missing config
	ExConfig ExitCode = 78
)
