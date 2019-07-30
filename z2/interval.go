package z2

import (
	"fmt"
	"math"

	"github.com/golang/geo/r1"
)

type Interval struct {
	Lo, Hi int
}

func (interval Interval) Eq(other Interval) bool {
	return interval.Lo == other.Lo && interval.Hi == other.Hi
}

// Contains is true if point is in the half-open interval,
// i.e. Lo is inclusive, Hi is exclusive
func (interval Interval) Contains(pt int) bool {
	return interval.Lo <= pt && pt < interval.Hi
}

func (interval Interval) Length() int {
	return interval.Hi - interval.Lo
}

func (interval Interval) String() string {
	return fmt.Sprintf("[Lo(%d), Hi(%d)]", interval.Lo, interval.Hi)
}

func snapRoundInterval(interval r1.Interval) Interval {
	return Interval{
		Lo: int(math.Floor(interval.Lo)),
		Hi: int(math.Floor(interval.Hi)),
	}
}
