package version

import (
	"fmt"
	"runtime"
)

// GitCommit returns the git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// BuildDate returns the date the binary was built
var BuildDate = ""

// Version returns the git tag that was compiled. This will be filled in by the compiler.
var Version = ""

// GoVersion returns the version of the go runtime used to compile the binary
var GoVersion = runtime.Version()

// OsArch returns the os and arch used to build the binary
var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
