// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package diff2

import (
    "fmt"
    _ "embed"
    )

//go:embed Version.dat
var Version string

func Hello() string {
    return fmt.Sprintf("Hello diff2 v%s", Version)
}
