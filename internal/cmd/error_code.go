package cmd

// ExitCode represents exit codes.
type ExitCode int

const (
	Ok            ExitCode = 0
	ExUsage       ExitCode = 64
	ExDataErr     ExitCode = 65
	ExNoInput     ExitCode = 66
	ExNoUser      ExitCode = 67
	ExNoHost      ExitCode = 68
	ExUnavailable ExitCode = 69
	ExSoftware    ExitCode = 70
	ExOsErr       ExitCode = 71
	ExOsFile      ExitCode = 72
	ExCantCreat   ExitCode = 73
	ExIOErr       ExitCode = 74
	ExTempFail    ExitCode = 75
	ExProtocol    ExitCode = 76
	ExNoPerm      ExitCode = 77
	ExConfig      ExitCode = 78
)
