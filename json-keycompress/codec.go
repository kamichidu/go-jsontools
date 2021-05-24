package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	DefaultKeysProperty = "@"

	DefaultValueProperty = "*"

	DefaultKeyAlternatives = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

type Encoder struct {
	once sync.Once

	keysProp string

	valueProp string

	keyAlts []string
}

func (e *Encoder) init() {
	e.once.Do(func() {
		if e.keysProp == "" {
			e.keysProp = DefaultKeysProperty
		}
		if e.valueProp == "" {
			e.valueProp = DefaultValueProperty
		}
		if len(e.keyAlts) == 0 {
			e.keyAlts = strings.Split(DefaultKeyAlternatives, "")
		}
	})
}

func (e *Encoder) Encode(data []byte) ([]byte, error) {
	e.init()
	if !json.Valid(data) {
		panic("invalid json data")
	}

	var v interface{}
	jd := json.NewDecoder(bytes.NewReader(data))
	jd.UseNumber()
	if err := jd.Decode(&v); err != nil {
		return nil, err
	}
	keys := map[string]string{}
	value := e.encode(keys, v)
	return json.Marshal(map[string]interface{}{
		e.keysProp:  keys,
		e.valueProp: value,
	})
}

func (e *Encoder) encodeKey(keys map[string]string, key string) string {
	if strings.HasPrefix(key, e.keysProp) {
		return key
	}
	for alt, k := range keys {
		if k == key {
			return e.keysProp + alt
		}
	}
	alt := e.keyAlts[len(keys)]
	keys[alt] = key
	return e.keysProp + alt
}

func (e *Encoder) encode(keys map[string]string, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return e.object(keys, v)
	case []interface{}:
		return e.array(keys, v)
	case string, json.Number, nil:
		return v
	default:
		panic(fmt.Sprintf("unhandled type %T", v))
	}
}

func (e *Encoder) object(keys map[string]string, value map[string]interface{}) interface{} {
	var l []string
	for k := range value {
		l = append(l, k)
	}
	sort.Strings(l)
	for _, k := range l {
		alt := e.encodeKey(keys, k)
		value[alt] = e.encode(keys, value[k])
		delete(value, k)
	}
	return value
}

func (e *Encoder) array(keys map[string]string, value []interface{}) interface{} {
	for i := range value {
		value[i] = e.encode(keys, value[i])
	}
	return value
}

type Decoder struct {
	once sync.Once

	keysProp string

	valueProp string
}

func (d *Decoder) init() {
	d.once.Do(func() {
		if d.keysProp == "" {
			d.keysProp = DefaultKeysProperty
		}
		if d.valueProp == "" {
			d.valueProp = DefaultValueProperty
		}
	})
}

func (d *Decoder) Decode(data []byte) ([]byte, error) {
	d.init()
	if !json.Valid(data) {
		panic("invalid json data")
	}

	var raw map[string]interface{}
	jd := json.NewDecoder(bytes.NewReader(data))
	jd.UseNumber()
	if err := jd.Decode(&raw); err != nil {
		return nil, err
	}
	var keys map[string]string
	if val, ok := raw[d.keysProp]; ok {
		var err error
		keys, err = d.decodeKeys(val)
		if err != nil {
			return nil, fmt.Errorf("failed to decode keys %q: %w", d.keysProp, err)
		}
	}
	value, _ := raw[d.valueProp]
	v := d.decode(keys, value)
	return json.Marshal(v)
}

func (d *Decoder) decodeKeys(keys interface{}) (map[string]string, error) {
	if keys == nil {
		return nil, nil
	}
	val, ok := keys.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid type of keys: %T", keys)
	}
	out := make(map[string]string, len(val))
	for k, v := range val {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type of key %q in keys: %T", k, v)
		}
		out[k] = s
	}
	return out, nil
}

func (d *Decoder) decode(keys map[string]string, value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return d.object(keys, v)
	case []interface{}:
		return d.array(keys, v)
	case string, json.Number, nil:
		return v
	default:
		panic(fmt.Sprintf("unhandled type %T", v))
	}
}

func (d *Decoder) object(keys map[string]string, value map[string]interface{}) interface{} {
	var l []string
	for k := range value {
		l = append(l, k)
	}
	for _, alt := range l {
		k := keys[strings.TrimPrefix(alt, "@")]
		value[k] = d.decode(keys, value[alt])
		delete(value, alt)
	}
	return value
}

func (d *Decoder) array(keys map[string]string, value []interface{}) interface{} {
	for i := range value {
		value[i] = d.decode(keys, value[i])
	}
	return value
}
