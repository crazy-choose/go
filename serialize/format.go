package serialize

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/text/encoding/unicode"
	"reflect"
	"strconv"
	"strings"
)

const (
	offsetTag     = "offset"
	lenTag        = "length"
	unicodeBinary = "unicodeBinary"
)

// Deserialize []byte -> Struct
func Deserialize(msg []byte, v interface{}, byteOffset uint16) error {
	var (
		offset uint64
		length uint64
		err    error
	)

	reflectValue := reflect.ValueOf(v).Elem()
	reflectType := reflectValue.Type()
	for i := 0; i < reflectType.NumField(); i++ {
		of := reflectType.Field(i).Tag.Get(offsetTag)
		if of == "" {
			continue
		}
		offset, err = strconv.ParseUint(of, 10, 64)
		offset = offset + uint64(byteOffset)
		if err != nil {
			return err
		}

		field := reflectValue.Field(i)
		if !field.CanInterface() {
			continue
		}

		getlenfunc := func() (uint64, error) {
			return strconv.ParseUint(reflectType.Field(i).Tag.Get(lenTag), 10, 64)
		}
		getUnicodeBinaryFunc := func() (uint64, error) {
			return strconv.ParseUint(reflectType.Field(i).Tag.Get(unicodeBinary), 10, 64)
		}

		switch field.Interface().(type) {
		case uint8: // 8 bit unsigned integer
			field.Set(reflect.ValueOf(msg[offset]))
		case uint16: // Little-Endian encoded 16 bit unsigned integer
			field.Set(reflect.ValueOf(binary.LittleEndian.Uint16(msg[offset : offset+2])))
		case uint32:
			field.Set(reflect.ValueOf(binary.LittleEndian.Uint32(msg[offset : offset+4])))
		case uint64:
			field.SetUint(binary.LittleEndian.Uint64(msg[offset : offset+8]))
		case int16: // Little-Endian encoded 16 bit signed integer
			field.Set(reflect.ValueOf(int16(binary.LittleEndian.Uint16(msg[offset : offset+2]))))
		case int32:
			field.Set(reflect.ValueOf(int32(binary.LittleEndian.Uint32(msg[offset : offset+4]))))
		case int64:
			field.Set(reflect.ValueOf(int64(binary.LittleEndian.Uint64(msg[offset : offset+8]))))
		case string: // ASCII characters which are left aligned and padded with spaces, unless otherwise specified
			if length, err = getlenfunc(); err == nil {
				if is, errr := getUnicodeBinaryFunc(); errr == nil && is != 0 {
					bs := msg[offset : offset+length]
					bs = bytes.Trim(bs, "\x00")
					bs, err = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder().Bytes(bs)
					if err != nil {
						continue
					}
					s := string(bs)
					s = strings.TrimSpace(s)
					field.Set(reflect.ValueOf(s))
					//field.SetString(str)
				} else {
					sb := string(msg[offset : offset+length])
					str := make([]rune, 0, len(sb))
					for _, value := range []rune(sb) {
						if value == 0 {
							continue
						}
						str = append(str, value)
					}
					field.SetString(string(str))
				}
			}
		}
	}
	return err
}

