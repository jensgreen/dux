// From du(1) man page:
// > The SIZE argument  is  an  integer  and  optional  unit  (example:  10K  is  10*1024).   Units  are
// > K,M,G,T,P,E,Z,Y (powers of 1024) or KB,MB,... (powers of 1000).
// We should diverge from that and display "K" etc for both SI and IEC for shorter strings.

package files

import "fmt"

const (
	/* TODO support --si option
	// SI unit prefixes
	K = 1000
	M = 1000 * 1000
	G = 1000 * 1000 * 1000
	T = 1000 * 1000 * 1000 * 1000
	P = 1000 * 1000 * 1000 * 1000 * 1000
	E = 1000 * 1000 * 1000 * 1000 * 1000 * 1000
	*/

	// IEC unit prefixes
	Ki = 1024
	Mi = 1024 * 1024
	Gi = 1024 * 1024 * 1024
	Ti = 1024 * 1024 * 1024 * 1024
	Pi = 1024 * 1024 * 1024 * 1024 * 1024
	Ei = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
)

func unitLetter(exp int) string {
	return []string{"K", "M", "G", "T", "P", "E"}[exp]
}

// HumanizeIEC returns a human-readable string with IEC unit prefix (1024 -> "1.0K")
func HumanizeIEC(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for n := n / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%3.1f%s", float64(n)/float64(div), unitLetter(exp))
}
