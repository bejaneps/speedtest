package measurement

import "fmt"

const (
	kb = 1000
	mb = 1000 * 1000
)

// BitRate represents measurement for download/upload speed
// it can be one of: bps(bits), Kbps (kilobits), Mbps (megabits)
type BitRate float64

// String returns prettified string representation
// of bit rate measurement in bps by default
func (b BitRate) String() string {
	return fmt.Sprintf("%.3f bps", b)
}

// MbpsStr returns prettified string representation
// of bit rate measurement in Megabits
func (b BitRate) MbpsStr() string {
	return fmt.Sprintf("%.3f Mbps", b/mb)
}

// KbpsStr returns prettified string representation
// of bit rate measurement in Kilobits
func (b BitRate) KbpsStr() string {
	return fmt.Sprintf("%.3f Kbps", b/kb)
}
