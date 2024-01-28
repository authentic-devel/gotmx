package gotmx

import "reflect"

// attr_iteration.go contains functions for iteration attributes:
// - g-outer-repeat: repeats the entire element for each item in a collection
// - g-inner-repeat: repeats the children for each item in a collection

// mapEntry represents a key-value pair when iterating over maps.
type mapEntry struct {
	Key   any
	Value any
}

// getOuterRepeat checks for the g-outer-repeat attribute and returns its value.
func getOuterRepeat(attrs AttributeMap) (modelPath string, found bool) {
	v, ok := attrs[attrDataGOuterRepeat]
	return v, ok
}

// getInnerRepeat checks for the g-inner-repeat attribute and returns its value.
func getInnerRepeat(attrs AttributeMap) (modelPath string, found bool) {
	v, ok := attrs[attrDataGInnerRepeat]
	return v, ok
}

// resolveIterable resolves a model path to an iterable value.
func resolveIterable(ctx *RenderContext, modelPath string, data any) (any, bool) {
	modelValue, isModelPath := ctx.ResolveValue(modelPath, data)
	if !isModelPath {
		return nil, false
	}
	return modelValue, true
}

// resolveAsSlice converts a value to a slice of items for iteration.
// Handles slices, arrays, and maps. For maps, each item is a mapEntry.
// For non-iterable values, returns a single-element slice containing the value.
func resolveAsSlice(modelValue any) []any {
	if modelValue == nil {
		return nil
	}

	// Fast path: common concrete types (no reflection)
	switch v := modelValue.(type) {
	case []any:
		return v
	case []string:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []int:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case []map[string]any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	case map[string]any:
		result := make([]any, 0, len(v))
		for k, val := range v {
			result = append(result, mapEntry{Key: k, Value: val})
		}
		return result
	}

	// Slow path: reflection for other types
	v := reflect.ValueOf(modelValue)
	kind := v.Kind()

	switch kind {
	case reflect.Map:
		result := make([]any, 0, v.Len())
		for _, k := range v.MapKeys() {
			key := k.Interface()
			value := v.MapIndex(k).Interface()
			result = append(result, mapEntry{Key: key, Value: value})
		}
		return result

	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result

	default:
		return []any{modelValue}
	}
}

// iterateValue provides a way to iterate over a value without creating an intermediate slice.
// The callback is called for each item. If callback returns an error, iteration stops.
// Context cancellation is checked before each iteration step to allow early exit
// when a client disconnects or a timeout is reached.
func iterateValue(ctx *RenderContext, modelValue any, callback func(item any) error) error {
	if modelValue == nil {
		return nil
	}

	// Fast path: common concrete types (no reflection)
	switch v := modelValue.(type) {
	case []any:
		for _, item := range v {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(item); err != nil {
				return err
			}
		}
		return nil
	case []string:
		for _, item := range v {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(item); err != nil {
				return err
			}
		}
		return nil
	case []int:
		for _, item := range v {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(item); err != nil {
				return err
			}
		}
		return nil
	case []map[string]any:
		for _, item := range v {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(item); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		for k, val := range v {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(mapEntry{Key: k, Value: val}); err != nil {
				return err
			}
		}
		return nil
	}

	// Slow path: reflection for other types
	v := reflect.ValueOf(modelValue)
	kind := v.Kind()

	switch kind {
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			key := k.Interface()
			value := v.MapIndex(k).Interface()
			if err := callback(mapEntry{Key: key, Value: value}); err != nil {
				return err
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := ctx.Context.Err(); err != nil {
				return err
			}
			if err := callback(v.Index(i).Interface()); err != nil {
				return err
			}
		}

	default:
		return callback(modelValue)
	}

	return nil
}
