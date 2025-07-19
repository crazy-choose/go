package misc

import (
	"fmt"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
	"math"
	"reflect"
	"strconv"
)

var (
	MaxFloat64    = -1.0
	MaxFloat64Dcm = decimal.NewFromFloat(MaxFloat64)
)

func PbToCSVReflect(msg proto.Message) ([]string, error) {
	var result []string
	val := reflect.ValueOf(msg).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		switch field.Kind() {
		case reflect.String:
			result = append(result, field.String())
		case reflect.Int, reflect.Int32, reflect.Int64:
			result = append(result, strconv.FormatInt(field.Int(), 10))
		case reflect.Float32, reflect.Float64:
			result = append(result, strconv.FormatFloat(field.Float(), 'f', -1, 64))
		case reflect.Bool:
			result = append(result, strconv.FormatBool(field.Bool()))
		default:
			// 处理其他类型或嵌套消息
			if m, ok := field.Interface().(proto.Message); ok {
				subFields, err := PbToCSVReflect(m)
				if err != nil {
					return nil, err
				}
				result = append(result, subFields...)
			} else {
				result = append(result, fmt.Sprintf("%v", field.Interface()))
			}
		}
	}

	return result, nil
}

// StructToSlice 方法将 Any 结构体转换为 []string
func StructToSlice(p any, hasMax bool) []string {
	// 取得结构体的反射类型
	val := reflect.ValueOf(p)
	typ := reflect.TypeOf(p)

	// 创建一个 []string 切片，用于存放转换后的值
	slice := make([]string, typ.NumField())

	// 遍历结构体的字段，并将字段的值转换为 string 存入切片
	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)
		switch field.Kind() {
		case reflect.String:
			slice[i] = field.String()
		case reflect.Int:
			slice[i] = fmt.Sprintf("%d", field.Int())
		case reflect.Bool:
			slice[i] = fmt.Sprintf("%t", field.Bool())
		case reflect.Float32, reflect.Float64:
			slice[i] = fmt.Sprintf("%f", field.Float())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			slice[i] = fmt.Sprintf("%d", field.Uint())
		case reflect.Complex64, reflect.Complex128:
			slice[i] = fmt.Sprintf("%v", field.Complex())
		case reflect.Struct:
			if field.Type() == reflect.TypeOf(decimal.Decimal{}) {
				if !hasMax {
					slice[i] = field.Interface().(decimal.Decimal).String()
				} else {
					decimalValue := field.Interface().(decimal.Decimal)
					if decimalValue.GreaterThanOrEqual(decimal.NewFromFloat(math.MaxFloat64)) {
						slice[i] = "-1.00"
					} else {
						slice[i] = field.Interface().(decimal.Decimal).String()
					}
				}
			} else {
				fmt.Println("[StructToSlice]Unsupported struct type")
			}
		default:
			//panic("Unsupported type")
			fmt.Println("[StructToSlice]Unsupported struct type")
		}
	}

	return slice
}

func SliceToStruct(slice []string, p any) error {
	val := reflect.ValueOf(p)
	typ := reflect.TypeOf(p)

	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("SliceToStruct: expects a pointer to struct")
	}

	val = val.Elem()
	typ = typ.Elem()

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("SliceToStruct: expects a pointer to struct")
	}

	if len(slice) != val.NumField() {
		return fmt.Errorf("SliceToStruct: slice length does not match struct fields")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i).Type

		switch fieldType.Kind() {
		case reflect.String:
			field.SetString(slice[i])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal, err := strconv.ParseInt(slice[i], 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal, err := strconv.ParseUint(slice[i], 10, 64)
			if err != nil {
				return err
			}
			field.SetUint(uintVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(slice[i])
			if err != nil {
				return err
			}
			field.SetBool(boolVal)
		case reflect.Float32, reflect.Float64:
			floatVal, err := strconv.ParseFloat(slice[i], 64)
			if err != nil {
				return err
			}
			field.SetFloat(floatVal)
		case reflect.Struct:
			if fieldType == reflect.TypeOf(decimal.Decimal{}) {
				decVal, err := decimal.NewFromString(slice[i])
				if err != nil {
					fmt.Println("[SliceToStruct]decimal.NewFromString error")
					return err
				}
				if decVal.Equal(decimal.NewFromFloat(math.MaxFloat64)) {
					field.Set(reflect.ValueOf(MaxFloat64Dcm))
				} else {
					field.Set(reflect.ValueOf(decVal))
				}
			} else {
				return fmt.Errorf("SliceToStruct: unsupported struct type")
			}
		default:
			return fmt.Errorf("SliceToStruct: unsupported type")
		}
	}

	return nil
}
