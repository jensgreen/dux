package z1

import (
	"math"

	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/geo/r1"
)

type Interval = geo.Interval[int]

func SnapRoundInterval(interval r1.Interval) Interval {
	return Interval{
		Lo: int(math.Floor(interval.Lo)),
		Hi: int(math.Floor(interval.Hi)),
	}
}
