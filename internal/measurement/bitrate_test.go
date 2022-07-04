package measurement

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	rate := BitRate(1051254.4242)

	rateStr := rate.String()
	assert.Equal(t, "1051254.424 bps", rateStr)
}

func TestMbpsStr(t *testing.T) {
	rate := BitRate(1051254.4242)

	rateMbpsStr := rate.MbpsStr()
	assert.Equal(t, "1.051 Mbps", rateMbpsStr)
}

func TestKbpsStr(t *testing.T) {
	rate := BitRate(1051254.4242)

	rateKbpsStr := rate.KbpsStr()
	assert.Equal(t, "1051.254 Kbps", rateKbpsStr)
}
