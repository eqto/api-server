package api

import (
	"strconv"
)

//Parameter ...
type Parameter struct {
	value interface{}
}

func (p *Parameter) set(value interface{}) {
	switch val := value.(type) {
	case int8:
		value = int(val)
	case int16:
		value = int(val)
	case int32:
		value = int(val)
	case int64:
		value = int(val)
	case uint8:
		value = uint(val)
	case uint16:
		value = uint(val)
	case uint32:
		value = uint(val)
	case uint64:
		value = uint(val)
	case float32:
		value = float64(val)
	default:
	}
	p.value = value
}

//Int ...
func (p *Parameter) Int() int {
	switch val := p.value.(type) {
	case int:
		return val
	case uint:
		return int(val)
	case string:
		if i, e := strconv.Atoi(val); e == nil {
			return i
		}
	case float64:
		return int(val)
	}
	return 0
}

func (p *Parameter) String() string {
	switch val := p.value.(type) {
	case int:
		return strconv.Itoa(val)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	}
	return ``
}

//Float ...
func (p *Parameter) Float() float64 {
	switch val := p.value.(type) {
	case int:
		return float64(val)
	case uint:
		return float64(val)
	case string:
		if f, e := strconv.ParseFloat(val, 64); e == nil {
			return f
		}
	case float64:
		return val
	}
	return 0
}
