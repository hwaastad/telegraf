package divergent

import (
	"log"
	"math"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/aggregators"
)

type Divergent struct {
	Stats []string `toml:"divergent"`

  cache map[uint64]aggregate
}

type aggregate struct {
	fields map[string]divergent
	name   string
	tags   map[string]string
}

type divergent struct {
	min   float64
  midTs time.Time
	max   float64
	maxTs time.Time
}

func NewDivergent() telegraf.Aggregator {
  m := &Divergent{}
	m.Reset()
	return m
}

var sampleConfig = `
  ## period is the flush & clear interval of the aggregator.
  period = "30s"
  ## If true drop_original will drop the original metrics and
  ## only send aggregates.
  drop_original = false
`

func (m *Divergent) SampleConfig() string {
	return sampleConfig
}

func (m *Divergent) Description() string {
	return "Keep the aggregate divergent of each metric passing through (counters)"
}

func (m *Divergent) Add(in telegraf.Metric) {
  id := in.HashID()
  if _, ok := m.cache[id]; !ok {
    a := aggregate{
      name:   in.Name(),
			tags:   in.Tags(),
			fields: make(map[string]divergent),
    }
    for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
				a.fields[field.Key] = divergent{
					min:   fv,
					max:   fv,
					minTs:  in.Time(),
					maxTs: in.Time(),
				}
			}
      m.cache[id] = a
  } else {
    for _, field := range in.FieldList() {
			if fv, ok := convert(field.Value); ok {
        if _, ok := m.cache[id].fields[field.Key]; !ok {
          // hit an uncached field of a cached metric
					m.cache[id].fields[field.Key] = divergent{
            min:   fv,
  					max:   fv,
  					minTs:  in.Time(),
  					maxTs: in.Time(),
					}
					continue
        }
        tmp := m.cache[id].fields[field.Key]
        tmp.max = fv
        tmp.maxTs = in.Time()

        m.cache[id].fields[field.Key] = tmp
      }
    }
  }
}

func (m *Divergent) Push(acc telegraf.Accumulator) {
  for key, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {
      diff := v.maxTs.Sub(v.minTs).Seconds()
      fields[k+"_divergent"] = (v.max - v.min)/diff
      m.cache[key].fields[k].min = m.cache[key].fields[k].max
        m.cache[key].fields[k].minTS = m.cache[key].fields[k].maxTs
    }
    if len(fields) > 0 {
      acc.AddFields(aggregate.name, fields, aggregate.tags)
    }
  }
}

func (m *Divergent) Reset() {
  	m.cache = make(map[uint64]aggregate)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func init() {
	aggregators.Add("divergent", func() telegraf.Aggregator {
		return NewDivergent()
	})
}
