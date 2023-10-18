package rpc

import "strconv"

type Parameter []string

func ParamList(values ...interface{}) Parameter {
	param := Parameter{}
	for _, value := range values {
		param.Add(value)
	}
	return param
}

func (param Parameter) Add(value interface{}) Parameter {
	switch value.(type) {
	case int:
		value = strconv.Itoa(value.(int))
	case int8:
		value = strconv.FormatInt(int64(value.(int8)), 10)
	case int16:
		value = strconv.FormatInt(int64(value.(int16)), 10)
	case int32:
		value = strconv.FormatInt(int64(value.(int32)), 10)
	case int64:
		value = strconv.FormatInt(value.(int64), 10)
	case uint:
		value = strconv.FormatUint(uint64(value.(uint)), 10)
	case uint8:
		value = strconv.FormatUint(uint64(value.(uint8)), 10)
	case uint16:
		value = strconv.FormatUint(uint64(value.(uint16)), 10)
	case uint32:
		value = strconv.FormatUint(uint64(value.(uint32)), 10)
	case uint64:
		value = strconv.FormatUint(value.(uint64), 10)
	}
	param = append(param, value.(string))
	return param
}
