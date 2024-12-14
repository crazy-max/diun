package parser

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/crazy-max/gonfig/types"
)

const defaultRawSliceSeparator = ","

type initializer interface {
	SetDefaults()
}

// FillerOpts Options for the filler.
type FillerOpts struct {
	AllowSliceAsStruct bool
	RawSliceSeparator  string
}

// Fill populates the fields of the element using the information in node.
func Fill(element interface{}, node *Node, opts FillerOpts) error {
	return newFiller(opts).Fill(element, node)
}

type filler struct {
	FillerOpts
}

func newFiller(opts FillerOpts) filler {
	if opts.RawSliceSeparator == "" {
		opts.RawSliceSeparator = defaultRawSliceSeparator
	}

	return filler{FillerOpts: opts}
}

// Fill populates the fields of the element using the information in node.
func (f filler) Fill(element interface{}, node *Node) error {
	if element == nil || node == nil {
		return nil
	}

	if node.Kind == 0 {
		return fmt.Errorf("missing node type: %s", node.Name)
	}

	root := reflect.ValueOf(element)
	if root.Kind() == reflect.Struct {
		return fmt.Errorf("struct are not supported, use pointer instead")
	}

	return f.fill(root.Elem(), node)
}

func (f filler) fill(field reflect.Value, node *Node) error {
	// related to allow-empty or ignore tag
	if node.Disabled {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(node.Value)
		return nil
	case reflect.Bool:
		val, err := strconv.ParseBool(node.Value)
		if err != nil {
			return err
		}
		field.SetBool(val)
		return nil
	case reflect.Int8:
		return setInt(field, node.Value, 8)
	case reflect.Int16:
		return setInt(field, node.Value, 16)
	case reflect.Int32:
		return setInt(field, node.Value, 32)
	case reflect.Int64, reflect.Int:
		return setInt(field, node.Value, 64)
	case reflect.Uint8:
		return setUint(field, node.Value, 8)
	case reflect.Uint16:
		return setUint(field, node.Value, 16)
	case reflect.Uint32:
		return setUint(field, node.Value, 32)
	case reflect.Uint64, reflect.Uint:
		return setUint(field, node.Value, 64)
	case reflect.Float32:
		return setFloat(field, node.Value, 32)
	case reflect.Float64:
		return setFloat(field, node.Value, 64)
	case reflect.Struct:
		return f.setStruct(field, node)
	case reflect.Pointer:
		return f.setPtr(field, node)
	case reflect.Map:
		return f.setMap(field, node)
	case reflect.Slice:
		return f.setSlice(field, node)
	default:
		return nil
	}
}

func (f filler) setPtr(field reflect.Value, node *Node) error {
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))

		if field.Type().Implements(reflect.TypeOf((*initializer)(nil)).Elem()) {
			method := field.MethodByName("SetDefaults")
			if method.IsValid() {
				method.Call([]reflect.Value{})
			}
		}
	}

	return f.fill(field.Elem(), node)
}

func (f filler) setStruct(field reflect.Value, node *Node) error {
	for _, child := range node.Children {
		fd := field.FieldByName(child.FieldName)

		zeroValue := reflect.Value{}
		if fd == zeroValue {
			return fmt.Errorf("field not found, node: %s (%s)", child.Name, child.FieldName)
		}

		err := f.fill(fd, child)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f filler) setSlice(field reflect.Value, node *Node) error {
	if field.Type().Elem().Kind() == reflect.Struct ||
		field.Type().Elem().Kind() == reflect.Pointer && field.Type().Elem().Elem().Kind() == reflect.Struct {
		return f.setSliceStruct(field, node)
	}

	if len(node.Value) == 0 {
		return nil
	}

	values := strings.Split(node.Value, f.RawSliceSeparator)
	if f.RawSliceSeparator != defaultRawSliceSeparator {
		if len(values) < 2 {
			// TODO(ldez): must be changed to an error.
			return makeSlice(field, strings.Split(node.Value, defaultRawSliceSeparator))
		}

		// TODO(ldez): this is related to raw map and file. Rethink the node parser.
		values = values[2:]
	}

	return makeSlice(field, values)
}

func makeSlice(field reflect.Value, values []string) error {
	slice := reflect.MakeSlice(field.Type(), len(values), len(values))
	field.Set(slice)

	for i := 0; i < len(values); i++ {
		value := strings.TrimSpace(values[i])

		switch field.Type().Elem().Kind() {
		case reflect.String:
			field.Index(i).SetString(value)
		case reflect.Int:
			val, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.Index(i).SetInt(val)
		case reflect.Int8:
			err := setInt(field.Index(i), value, 8)
			if err != nil {
				return err
			}
		case reflect.Int16:
			err := setInt(field.Index(i), value, 16)
			if err != nil {
				return err
			}
		case reflect.Int32:
			err := setInt(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Int64:
			err := setInt(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Uint:
			val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			field.Index(i).SetUint(val)
		case reflect.Uint8:
			err := setUint(field.Index(i), value, 8)
			if err != nil {
				return err
			}
		case reflect.Uint16:
			err := setUint(field.Index(i), value, 16)
			if err != nil {
				return err
			}
		case reflect.Uint32:
			err := setUint(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Uint64:
			err := setUint(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Float32:
			err := setFloat(field.Index(i), value, 32)
			if err != nil {
				return err
			}
		case reflect.Float64:
			err := setFloat(field.Index(i), value, 64)
			if err != nil {
				return err
			}
		case reflect.Bool:
			val, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.Index(i).SetBool(val)
		default:
			return fmt.Errorf("unsupported type: %s", field.Type().Elem())
		}
	}
	return nil
}

func (f filler) setSliceStruct(field reflect.Value, node *Node) error {
	if f.AllowSliceAsStruct && node.Tag.Get(TagLabelSliceAsStruct) != "" {
		return f.setSliceAsStruct(field, node)
	}

	field.Set(reflect.MakeSlice(field.Type(), len(node.Children), len(node.Children)))

	for i, child := range node.Children {
		// use Ptr to allow "SetDefaults"
		value := reflect.New(reflect.PointerTo(field.Type().Elem()))
		err := f.setPtr(value, child)
		if err != nil {
			return err
		}

		field.Index(i).Set(value.Elem().Elem())
	}

	return nil
}

func (f filler) setSliceAsStruct(field reflect.Value, node *Node) error {
	if len(node.Children) == 0 {
		return fmt.Errorf("invalid slice: node %s", node.Name)
	}

	// use Ptr to allow "SetDefaults"
	value := reflect.New(reflect.PointerTo(field.Type().Elem()))
	if err := f.setPtr(value, node); err != nil {
		return err
	}

	elem := value.Elem().Elem()

	field.Set(reflect.MakeSlice(field.Type(), 1, 1))
	field.Index(0).Set(elem)

	return nil
}

func (f filler) setMap(field reflect.Value, node *Node) error {
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	if field.Type().Elem().Kind() == reflect.Interface {
		err := f.fillRawValue(field, node, false)
		if err != nil {
			return err
		}

		for _, child := range node.Children {
			err = f.fillRawValue(field, child, true)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, child := range node.Children {
		ptrValue := reflect.New(reflect.PointerTo(field.Type().Elem()))

		err := f.fill(ptrValue, child)
		if err != nil {
			return err
		}

		value := ptrValue.Elem().Elem()

		key := reflect.ValueOf(child.Name)
		field.SetMapIndex(key, value)
	}

	return nil
}

func setInt(field reflect.Value, value string, bitSize int) error {
	switch field.Type() {
	case reflect.TypeOf(types.Duration(0)):
		return setDuration(field, value, bitSize, time.Second)
	case reflect.TypeOf(time.Duration(0)):
		return setDuration(field, value, bitSize, time.Nanosecond)
	default:
		val, err := strconv.ParseInt(value, 10, bitSize)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(val).Convert(field.Type()))
		return nil
	}
}

func setDuration(field reflect.Value, value string, bitSize int, defaultUnit time.Duration) error {
	val, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.Set(reflect.ValueOf(time.Duration(val) * defaultUnit).Convert(field.Type()))
		return nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(duration).Convert(field.Type()))
	return nil
}

func setUint(field reflect.Value, value string, bitSize int) error {
	val, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(val).Convert(field.Type()))
	return nil
}

func setFloat(field reflect.Value, value string, bitSize int) error {
	val, err := strconv.ParseFloat(value, bitSize)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(val).Convert(field.Type()))
	return nil
}

func (f filler) fillRawValue(field reflect.Value, node *Node, subMap bool) error {
	m, ok := node.RawValue.(map[string]interface{})
	if !ok {
		return nil
	}

	if _, self := m[node.Name]; self || !subMap {
		for k, v := range m {
			if f.RawSliceSeparator == defaultRawSliceSeparator {
				field.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
				continue
			}

			// TODO(ldez): all the next section is related to raw map and file. Rethink the node parser.

			s, ok := v.(string)
			if !ok || len(s) == 0 || !strings.HasPrefix(s, f.RawSliceSeparator) {
				rawValue, err := f.cleanRawValue(reflect.ValueOf(v))
				if err != nil {
					return err
				}

				field.SetMapIndex(reflect.ValueOf(k), rawValue)
				continue
			}

			// typed slice

			value, err := f.fillRawTypedSlice(s)
			if err != nil {
				return err
			}

			field.SetMapIndex(reflect.ValueOf(k), value)
		}

		return nil
	}

	// In the case of sub-map, fill raw typed slices recursively.
	_, err := f.fillRawMapWithTypedSlice(m)
	if err != nil {
		return err
	}

	p := map[string]interface{}{node.Name: m}
	node.RawValue = p

	field.SetMapIndex(reflect.ValueOf(node.Name), reflect.ValueOf(p[node.Name]))

	return nil
}

func (f filler) fillRawMapWithTypedSlice(elt interface{}) (reflect.Value, error) {
	eltValue := reflect.ValueOf(elt)

	switch eltValue.Kind() {
	case reflect.String:
		if strings.HasPrefix(elt.(string), f.RawSliceSeparator) {
			sliceValue, err := f.fillRawTypedSlice(elt.(string))
			if err != nil {
				return eltValue, err
			}

			return sliceValue, nil
		}

	case reflect.Map:
		for k, v := range elt.(map[string]interface{}) {
			value, err := f.fillRawMapWithTypedSlice(v)
			if err != nil {
				return eltValue, err
			}

			eltValue.SetMapIndex(reflect.ValueOf(k), value)
		}
	}

	return eltValue, nil
}

func (f filler) fillRawTypedSlice(s string) (reflect.Value, error) {
	raw := strings.Split(s, f.RawSliceSeparator)

	rawType, err := strconv.Atoi(raw[1])
	if err != nil {
		return reflect.Value{}, err
	}

	kind := reflect.Kind(rawType)

	slice := reflect.MakeSlice(reflect.TypeOf([]interface{}{}), len(raw[2:]), len(raw[2:]))

	for i := 0; i < len(raw[2:]); i++ {
		switch kind {
		case reflect.Bool:
			val, err := strconv.ParseBool(raw[i+2])
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse bool: %s, %w", raw[i+2], err)
			}
			slice.Index(i).Set(reflect.ValueOf(val))
		case reflect.Int:
			val, err := strconv.ParseInt(raw[i+2], 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse int: %s, %w", raw[i+2], err)
			}
			slice.Index(i).Set(reflect.ValueOf(val))
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := strconv.ParseInt(raw[i+2], 10, int(math.Pow(2, float64(kind-reflect.Int+2))))
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse %s: %s, %w", kind, raw[i+2], err)
			}
			slice.Index(i).Set(reflect.ValueOf(val))
		case reflect.Uint:
			val, err := strconv.ParseUint(raw[i+2], 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse uint: %s, %w", raw[i+2], err)
			}
			slice.Index(i).Set(reflect.ValueOf(val))
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, err := strconv.ParseUint(raw[i+2], 10, int(math.Pow(2, float64(kind-reflect.Uint+2))))
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse uint: %s, %w", raw[i+2], err)
			}
			slice.Index(i).Set(reflect.ValueOf(val))
		case reflect.Float32:
			err := setFloat(slice.Index(i), raw[i+2], 32)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse float32: %s, %w", raw[i+2], err)
			}
		case reflect.Float64:
			err := setFloat(slice.Index(i), raw[i+2], 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("parse float64: %s, %w", raw[i+2], err)
			}
		case reflect.String:
			slice.Index(i).Set(reflect.ValueOf(raw[i+2]))
		default:
			return reflect.Value{}, fmt.Errorf("unsupported kind: %d", kind)
		}
	}

	return slice, nil
}

func (f filler) cleanRawValue(value reflect.Value) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Pointer:
		rawValue, err := f.cleanRawValue(value.Elem())
		if err != nil {
			return reflect.Value{}, err
		}

		value.Elem().Set(rawValue)

	case reflect.Map:
		keys := value.MapKeys()
		for _, key := range keys {
			v := value.MapIndex(key)

			rawValue, err := f.cleanRawValue(v)
			if err != nil {
				return reflect.Value{}, err
			}

			value.SetMapIndex(key, rawValue)
		}

	case reflect.Slice:
		if value.IsZero() {
			return value, nil
		}

		for i := 0; i < value.Len(); i++ {
			if !value.Index(i).IsZero() {
				rawValue, err := f.cleanRawValue(value.Index(i))
				if err != nil {
					return reflect.Value{}, err
				}

				value.Index(i).Set(rawValue)
			}
		}

	case reflect.Interface:
		return f.cleanRawValue(value.Elem())

	case reflect.String:
		if strings.HasPrefix(value.String(), f.RawSliceSeparator) {
			return f.fillRawTypedSlice(value.String())
		}
	}

	return value, nil
}
