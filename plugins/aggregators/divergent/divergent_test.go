package divergent

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var t = time.Now()

var m1, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"ifInOctets": 200,
		"ifOutOctets": 10,
	},
	t,
)

var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"ifInOctets": 400,
		"ifOutOctets": 10,
	},
	t.Add(time.Second * 10),
)

func BenchmarkApply(b *testing.B) {
	vc := NewDivergent()

	for n := 0; n < b.N; n++ {
		vc.Add(m1)
		vc.Add(m2)
	}
}

func TestMinMaxWithPeriod(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewDivergent()

	minmax.Add(m1)
	minmax.Add(m2)
	minmax.Push(&acc)

	expectedFields := map[string]interface{}{
		"ifOutOctets_divergent": float64(0),
		"ifInOctets_divergent": float64(20),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}
