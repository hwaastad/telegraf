package divergent

import (

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
	min    float64
  min_ts int64
	max    float64
	max_ts int64
}

func NewDivergent() telegraf.Aggregator {
  m := &Divergent{}
	m.cache = make(map[uint64]aggregate)
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
					min_ts:  in.Time().Unix(),
					max_ts: in.Time().Unix(),
				}
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
  					min_ts:  in.Time().Unix(),
  					max_ts: in.Time().Unix(),
					}
					continue
        }
        tmp := m.cache[id].fields[field.Key]
        tmp.max = fv
        tmp.max_ts = in.Time().Unix()
        m.cache[id].fields[field.Key] = tmp
      }
    }
  }
}

func (m *Divergent) Push(acc telegraf.Accumulator) {
  for _, aggregate := range m.cache {
		fields := map[string]interface{}{}
		for k, v := range aggregate.fields {
      diff := v.max_ts - v.min_ts
      tmp := aggregate.fields[k];
      fields[k+"_divergent"] = (v.max - v.min)/float64(diff)
      tmp.min = v.max
      tmp.min_ts = v.max_ts
      aggregate.fields[k] = tmp
    }
    if len(fields) > 0 {
      acc.AddFields(aggregate.name, fields, aggregate.tags)
    }
  }
}

func (m *Divergent) Reset() {
  	//m.cache = make(map[uint64]aggregate)
}

func convert(in interface{}) (float64, bool) {
	switch v := in.(type) {
  case float64:
		return v, true
	case int64:
		return float64(v), true
	case uint64:
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
