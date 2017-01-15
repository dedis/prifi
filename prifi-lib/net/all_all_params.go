package net

//A wrapper around the interface{} type, since protobuf can't handle map[]interface{}
type Interface struct {
	Data interface{}
}

// ALL_ALL_PARAMETERS message contains all the parameters used by the protocol.
type ALL_ALL_PARAMETERS_NEW struct {
	ForceParams bool
	Params      map[string]Interface
}

//A builder for ALL_ALL_PARAMETERS messages. You can call Add(key, val) and BuildMessage()
type ALL_ALL_PARAMETERS_BUILDER struct {
	data map[string]Interface
}

/**
 * Creates a builder for ALL_ALL_PARAMETERS
 */
func NewALL_ALL_PARAMETERS_BUILDER() *ALL_ALL_PARAMETERS_BUILDER {
	builder := ALL_ALL_PARAMETERS_BUILDER{}
	builder.data = make(map[string]Interface)
	return &builder
}

/**
 * Adds a (key, val) to the ALL_ALL_PARAMS message builder
 */
func (b *ALL_ALL_PARAMETERS_BUILDER) Add(key string, val interface{}) {
	b.data[key] = Interface{Data: val}
}

/**
 * Creates a ALL_ALL_PARAMETERS message
 */
func (b *ALL_ALL_PARAMETERS_BUILDER) BuildMessage(forceParams bool) *ALL_ALL_PARAMETERS_NEW {
	msg := ALL_ALL_PARAMETERS_NEW{Params: b.data, ForceParams: forceParams}
	return &msg
}

/**
 * From the message, returns the "data[key]" if it exists, or "elseVal"
 */
func (m *ALL_ALL_PARAMETERS_NEW) ValueOrElse(key string, elseVal interface{}) interface{} {
	if val, ok := m.Params[key]; ok {
		return val.Data
	}
	return elseVal
}

/**
 * From the message, returns the "data[key]" if it exists, or "elseVal"
 */
func (m *ALL_ALL_PARAMETERS_NEW) BoolValueOrElse(key string, elseVal bool) bool {
	if val, ok := m.Params[key]; ok {
		return val.Data.(bool)
	}
	return elseVal
}

/**
 * From the message, returns the "data[key]" if it exists, or "elseVal"
 */
func (m *ALL_ALL_PARAMETERS_NEW) IntValueOrElse(key string, elseVal int) int {
	if val, ok := m.Params[key]; ok {
		return val.Data.(int)
	}
	return elseVal
}

/**
 * From the message, returns the "data[key]" if it exists, or "elseVal"
 */
func (m *ALL_ALL_PARAMETERS_NEW) StringValueOrElse(key string, elseVal string) string {
	if val, ok := m.Params[key]; ok {
		return val.Data.(string)
	}
	return elseVal
}