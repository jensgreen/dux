package z2

import (
	"math"

	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/r1"
)

type Interval = geo.Interval[int]

func snapRoundInterval(interval r1.Interval) Interval {
	return Interval{
		Lo: int(math.Floor(interval.Lo)),
		Hi: int(math.Floor(interval.Hi)),
	}
}
