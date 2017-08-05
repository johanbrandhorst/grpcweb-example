package recoverer

import (
	"fmt"

	"github.com/rusco/qunit"
)

func Recover() {
	e := recover()
	if e == nil {
		return
	}

	qunit.Ok(false, fmt.Sprintf("Saw panic: %v", e))
}
