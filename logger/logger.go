package logger

import (
	"os"

	"github.com/op/go-logging"
)

// Logging for keyserver:
//   - Generic requests = Info
//   - Active key response = Notice
//   - Key active/inactive change = Notice
//   - Inactive key response = Warn

var Log = logging.MustGetLogger("keyserver")

// Init sets up the logger
func Init() {
	// color for stdout
	colorFormat := logging.MustStringFormatter(
		`%{color}%{time:02/Jan/2006 15:04:05} - %{message}%{color:reset}`,
	)
	// no color for file output
	plainFormat := logging.MustStringFormatter(
		`%{time:02/Jan/2006 15:04:05} - %{message}`,
	)

	logFile, err := os.OpenFile("./keyserver.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	fileBackendRaw := logging.NewLogBackend(logFile, "", 0)
	fileBackend := logging.NewBackendFormatter(fileBackendRaw, plainFormat)

	stdoutBackendRaw := logging.NewLogBackend(os.Stdout, "", 0)
	stdoutBackend := logging.NewBackendFormatter(stdoutBackendRaw, colorFormat)

	// Start by only showing notice and higher with stdout
	stdoutLeveled := logging.AddModuleLevel(stdoutBackend)
	stdoutLeveled.SetLevel(logging.NOTICE, "")

	logging.SetBackend(stdoutLeveled, fileBackend)
}
