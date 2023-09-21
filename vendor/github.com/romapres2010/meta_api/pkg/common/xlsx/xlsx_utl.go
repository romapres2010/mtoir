package xlsx

import (
	"reflect"
	"time"

	"gopkg.in/guregu/null.v4"
)

//type T interface{ t() }
//
//// nolint gochecknoglobals
//var (
//    tType    = reflect.TypeOf((*T)(nil)).Elem()
//    timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
//)

//func getCellName(colIndex int) string {
//	name := make([]byte, 0, 3) // max 16,384 columns (2022)
//	const aLen = 'Z' - 'A' + 1 // alphabet length
//	for ; colIndex > 0; colIndex /= aLen + 1 {
//		name = append(name, byte('A'+(colIndex-1)%aLen))
//	}
//	for i, j := 0, len(name)-1; i < j; i, j = i+1, j-1 {
//		name[i], name[j] = name[j], name[i]
//	}
//	return string(name)
//}

func toFloat64(v interface{}) (float64, bool) {
	switch fv := v.(type) {
	case int:
		return float64(fv), true
	case int8:
		return float64(fv), true
	case int16:
		return float64(fv), true
	case int32:
		return float64(fv), true
	case int64:
		return float64(fv), true
	case uint:
		return float64(fv), true
	case uint8:
		return float64(fv), true
	case uint16:
		return float64(fv), true
	case uint32:
		return float64(fv), true
	case uint64:
		return float64(fv), true
	case float32:
		return float64(fv), true
	case float64:
		return fv, true
	}

	return 0, false
}

func findSheetName(value any) string {
	t := reflect.TypeOf(value)
	for i := 0; i < t.NumField(); i++ {
		if sheetName := t.Field(i).Tag.Get(XLSX_TAG_SHEET); sheetName != "" {
			return sheetName
		}
	}
	return ""
}

func IsNotStructType(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
		switch fieldType {
		case reflect.TypeOf((*null.String)(nil)).Elem(), reflect.TypeOf((*null.Float)(nil)).Elem(), reflect.TypeOf((*null.Int)(nil)).Elem(), reflect.TypeOf((*null.Bool)(nil)).Elem(), reflect.TypeOf((*null.Time)(nil)).Elem():
			return true
		}
		return false
	} else {
		return true
	}
}

func isNotStruct(value any) bool {
	fieldType := reflect.TypeOf(value)
	if fieldType.Kind() == reflect.Struct || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Struct) {
		switch value.(type) {
		case null.String, null.Float, null.Int, null.Bool, null.Time, time.Time:
			return true
		}
		return false
	} else {
		return true
	}
}

func IsSlice(value any) bool {
	fieldType := reflect.TypeOf(value)
	if fieldType.Kind() == reflect.Slice || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Slice) {
		return true
	}
	return false
}

func IsSliceType(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.Slice || (fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Slice) {
		return true
	} else {
		return false
	}
}
