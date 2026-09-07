package maybe

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
)

type T[Inner any] struct {
	value   Inner
	present bool
}

func None[Inner any]() T[Inner] {
	return T[Inner]{}
}

func Some[Inner any](value Inner) T[Inner] {
	return T[Inner]{
		value:   value,
		present: true,
	}
}

func (m *T[Inner]) IsSome() bool {
	return m.present
}

func (m *T[Inner]) Get() (*Inner, bool) {
	if m.present {
		return &m.value, true
	}
	return nil, false
}

func (m *T[Inner]) Value() (Inner, bool) {
	if m.present {
		return m.value, true
	}
	var zero Inner
	return zero, false
}

func (m *T[Inner]) Or(defaultItem Inner) Inner {
	if m.present {
		return m.value
	}
	return defaultItem
}

func (m *T[Inner]) Unwrap() *Inner {
	if !m.present {
		panic("unwrapped Maybe with no value")
	}
	return &m.value
}

func (m *T[Inner]) UnwrapValue() Inner {
	if !m.present {
		panic("unwrapped Maybe with no value")
	}
	return m.value
}

func (m *T[Inner]) MarshalJSONTo(encoder *jsontext.Encoder) error {
	value, ok := m.Get()
	if !ok {
		omitZero, _ := json.GetOption(encoder.Options(), json.OmitZeroStructFields)
		if omitZero {
			// Don't write anything
		} else {
			encoder.WriteToken(jsontext.Null)
		}
		return nil
	}
	return json.MarshalEncode(encoder, value)
}

func IfNil[T any](pointer *T, defaultValue T) T {
	if pointer == nil {
		return defaultValue
	}
	return *pointer
}

func IfNilPtr[T any, TP *T](pointer, defaultPointer TP) TP {
	if pointer == nil {
		return defaultPointer
	}
	return pointer
}
