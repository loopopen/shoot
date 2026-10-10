package dest

import "time"

type Event struct {
	Elapsed time.Duration
}

type Label int32

type Record struct {
	Label Label
}
