package ast

// ToValue converts an AST Document's value to a plain Go representation.
// Returns: nil, bool, int64, float64, string, []any, or *OrderedMapResult.
// This is used internally by the maml package.
func ToValue(doc *Document) any {
	return nodeToValue(doc.Value)
}

// OrderedMapResult represents an ordered map from AST conversion.
type OrderedMapResult struct {
	Keys   []string
	Values []any
}

func nodeToValue(node ValueNode) any {
	switch n := node.(type) {
	case *StringNode:
		return n.Value
	case *RawStringNode:
		return n.Value
	case *IntegerNode:
		return n.Value
	case *FloatNode:
		return n.Value
	case *BooleanNode:
		return n.Value
	case *NullNode:
		return nil
	case *ObjectNode:
		result := &OrderedMapResult{
			Keys:   make([]string, len(n.Properties)),
			Values: make([]any, len(n.Properties)),
		}
		for i, prop := range n.Properties {
			result.Keys[i] = prop.Key.KeyValue()
			result.Values[i] = nodeToValue(prop.Value)
		}
		return result
	case *ArrayNode:
		arr := make([]any, len(n.Elements))
		for i, el := range n.Elements {
			arr[i] = nodeToValue(el.Value)
		}
		return arr
	default:
		return nil
	}
}
