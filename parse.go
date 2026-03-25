package maml

import (
	"github.com/maml-dev/go-maml/ast"
)

// Parse parses a MAML source string into a Value.
func Parse(source string) (Value, error) {
	doc, err := ast.Parse(source)
	if err != nil {
		return Value{}, err
	}
	return convertToValue(ast.ToValue(doc)), nil
}

func convertToValue(v any) Value {
	switch val := v.(type) {
	case nil:
		return NewNull()
	case bool:
		return NewBool(val)
	case int64:
		return NewInt(val)
	case float64:
		return NewFloat(val)
	case string:
		return NewString(val)
	case []any:
		arr := make([]Value, len(val))
		for i, item := range val {
			arr[i] = convertToValue(item)
		}
		return NewArray(arr)
	case *ast.OrderedMapResult:
		m := NewOrderedMap()
		for i, key := range val.Keys {
			m.Set(key, convertToValue(val.Values[i]))
		}
		return NewObject(m)
	default:
		return NewNull()
	}
}
