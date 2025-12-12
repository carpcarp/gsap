package sap

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TypeCoercer handles type coercion
type TypeCoercer struct {
	visited map[string]bool // Track visited types for cycle detection
}

// NewTypeCoercer creates a new type coercer
func NewTypeCoercer() *TypeCoercer {
	return &TypeCoercer{
		visited: make(map[string]bool),
	}
}

// Coerce transforms a value to match the target type
func (c *TypeCoercer) Coerce(value interface{}, targetType reflect.Type) (interface{}, *Score, error) {
	if value == nil {
		return nil, &Score{total: 0}, nil
	}

	score := &Score{flags: make(map[string]int)}
	result, err := c.coerceValue(value, targetType, score)
	return result, score, err
}

func (c *TypeCoercer) coerceValue(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	// Handle nil
	if value == nil {
		return nil, nil
	}

	// Handle interface{} target
	if targetType.Kind() == reflect.Interface {
		return value, nil
	}

	// Get the actual type of the value
	valueType := reflect.TypeOf(value)

	// If types match, return as-is
	if valueType == targetType {
		return value, nil
	}

	// Handle pointers
	if targetType.Kind() == reflect.Ptr {
		// If value is nil, return nil pointer
		if value == nil {
			return reflect.Zero(targetType).Interface(), nil
		}
		elem, err := c.coerceValue(value, targetType.Elem(), score)
		if err != nil {
			return nil, err
		}
		// If the coerced element is nil, return nil pointer
		if elem == nil {
			return reflect.Zero(targetType).Interface(), nil
		}
		result := reflect.New(targetType.Elem())
		result.Elem().Set(reflect.ValueOf(elem))
		return result.Interface(), nil
	}

	// Handle basic types
	switch targetType.Kind() {
	case reflect.String:
		return c.coerceToString(value, score)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.coerceToInt(value, targetType, score)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.coerceToUint(value, targetType, score)
	case reflect.Float32, reflect.Float64:
		return c.coerceToFloat(value, targetType, score)
	case reflect.Bool:
		return c.coerceToBool(value, score)
	case reflect.Slice:
		return c.coerceToSlice(value, targetType, score)
	case reflect.Array:
		return c.coerceToArray(value, targetType, score)
	case reflect.Map:
		return c.coerceToMap(value, targetType, score)
	case reflect.Struct:
		return c.coerceToStruct(value, targetType, score)
	default:
		return nil, fmt.Errorf("unsupported type: %v", targetType)
	}
}

// coerceToString converts value to string
func (c *TypeCoercer) coerceToString(value interface{}, score *Score) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		// JSON numbers are float64
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10), nil
		}
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// coerceToInt converts value to integer
func (c *TypeCoercer) coerceToInt(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	var intVal int64

	switch v := value.(type) {
	case float64:
		intVal = int64(v)
		// Track if we rounded
		if float64(intVal) != v {
			score.AddFlag("FloatToInt", 1)
		}

	case string:
		// Try to parse as number
		parsed, err := parseNumber(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string to int: %v", err)
		}
		intVal = int64(parsed)
		score.AddFlag("StringToInt", 2)

	case bool:
		if v {
			intVal = 1
		}
		score.AddFlag("BoolToInt", 2)

	default:
		return nil, fmt.Errorf("cannot convert %T to int", value)
	}

	// Convert to target integer type
	result := reflect.ValueOf(intVal)
	return result.Convert(targetType).Interface(), nil
}

// coerceToUint converts value to unsigned integer
func (c *TypeCoercer) coerceToUint(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	intVal, err := c.coerceToInt(value, reflect.TypeOf(int64(0)), score)
	if err != nil {
		return nil, err
	}

	uintVal := uint64(intVal.(int64))
	result := reflect.ValueOf(uintVal)
	return result.Convert(targetType).Interface(), nil
}

// coerceToFloat converts value to float
func (c *TypeCoercer) coerceToFloat(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	var floatVal float64

	switch v := value.(type) {
	case float64:
		floatVal = v

	case string:
		parsed, err := parseNumber(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert string to float: %v", err)
		}
		floatVal = parsed
		score.AddFlag("StringToFloat", 2)

	case int, int8, int16, int32, int64:
		floatVal = float64(reflect.ValueOf(v).Int())

	default:
		return nil, fmt.Errorf("cannot convert %T to float", value)
	}

	result := reflect.ValueOf(floatVal)
	return result.Convert(targetType).Interface(), nil
}

// coerceToBool converts value to boolean
func (c *TypeCoercer) coerceToBool(value interface{}, score *Score) (interface{}, error) {
	switch v := value.(type) {
	case bool:
		return v, nil

	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		switch lower {
		case "true", "yes", "1", "on":
			score.AddFlag("StringToBool", 1)
			return true, nil
		case "false", "no", "0", "off":
			score.AddFlag("StringToBool", 1)
			return false, nil
		default:
			return nil, fmt.Errorf("cannot convert string to bool: %s", v)
		}

	case float64:
		score.AddFlag("NumberToBool", 1)
		return v != 0, nil

	default:
		return nil, fmt.Errorf("cannot convert %T to bool", value)
	}
}

// coerceToSlice converts value to slice
func (c *TypeCoercer) coerceToSlice(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	// Convert to []interface{} first
	var items []interface{}

	switch v := value.(type) {
	case []interface{}:
		items = v
	default:
		// Single item, wrap in slice
		items = []interface{}{value}
	}

	elemType := targetType.Elem()
	result := reflect.MakeSlice(targetType, len(items), len(items))

	for i, item := range items {
		elem, err := c.coerceValue(item, elemType, score)
		if err != nil {
			return nil, err
		}
		result.Index(i).Set(reflect.ValueOf(elem))
	}

	return result.Interface(), nil
}

// coerceToArray converts value to array
func (c *TypeCoercer) coerceToArray(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	// Similar to slice but with fixed size
	var items []interface{}

	switch v := value.(type) {
	case []interface{}:
		items = v
	default:
		items = []interface{}{value}
	}

	result := reflect.New(targetType).Elem()
	elemType := targetType.Elem()

	for i := 0; i < targetType.Len() && i < len(items); i++ {
		elem, err := c.coerceValue(items[i], elemType, score)
		if err != nil {
			return nil, err
		}
		result.Index(i).Set(reflect.ValueOf(elem))
	}

	return result.Interface(), nil
}

// coerceToMap converts value to map
func (c *TypeCoercer) coerceToMap(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	mapVal, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %T to map", value)
	}

	result := reflect.MakeMap(targetType)
	keyType := targetType.Key()
	elemType := targetType.Elem()

	for k, v := range mapVal {
		key, err := c.coerceValue(k, keyType, score)
		if err != nil {
			return nil, err
		}
		elem, err := c.coerceValue(v, elemType, score)
		if err != nil {
			return nil, err
		}
		result.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(elem))
	}

	return result.Interface(), nil
}

// coerceToStruct converts value to struct
func (c *TypeCoercer) coerceToStruct(value interface{}, targetType reflect.Type, score *Score) (interface{}, error) {
	mapVal, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot convert %T to struct", value)
	}

	result := reflect.New(targetType).Elem()

	// Get struct fields
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldType := field.Type

		// Find matching key in map
		var mapKey string
		var mapValue interface{}

		// Try JSON tag first
		if tag, ok := field.Tag.Lookup("json"); ok {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				if v, ok := mapVal[parts[0]]; ok {
					mapKey = parts[0]
					mapValue = v
				}
			}
		}

		// Try field name
		if mapKey == "" {
			if v, ok := mapVal[field.Name]; ok {
				mapKey = field.Name
				mapValue = v
			}
		}

		// Try case-insensitive match
		if mapKey == "" {
			for k, v := range mapVal {
				if strings.EqualFold(k, field.Name) {
					mapKey = k
					mapValue = v
					score.AddFlag("FuzzyFieldMatch", 1)
					break
				}
			}
		}

		// If found, coerce and set
		if mapKey != "" {
			elem, err := c.coerceValue(mapValue, fieldType, score)
			if err != nil {
				// Skip fields that fail to coerce if they're optional
				continue
			}
			// Handle nil values properly - use zero value for the type
			if elem == nil {
				result.Field(i).Set(reflect.Zero(fieldType))
			} else {
				result.Field(i).Set(reflect.ValueOf(elem))
			}
		}
	}

	return result.Interface(), nil
}

// parseNumber parses a string as a number
func parseNumber(s string) (float64, error) {
	s = strings.TrimSpace(s)

	// Remove currency symbols and commas
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")

	// Handle fractions like "1/5"
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) == 2 {
			num, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			denom, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err1 == nil && err2 == nil && denom != 0 {
				return num / denom, nil
			}
		}
	}

	return strconv.ParseFloat(s, 64)
}
