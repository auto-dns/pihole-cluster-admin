package healthcheck

import "github.com/rs/zerolog"

type Deps struct {
	Db     pinger
	Logger zerolog.Logger
}
