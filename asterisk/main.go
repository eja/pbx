// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package asterisk

import (
	"log/slog"
	"sync"
)

var log = sync.OnceValue(func() *slog.Logger {
	return slog.Default().With("app", "pbx", "pkg", "asterisk")
})
