package kdebug

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func VarDump(s interface{}) {
	indent := ""
	switch s.(type) {
	case reflect.Value:
		varDump(s.(reflect.Value), indent, "")
	default:
		varDump(reflect.ValueOf(s), indent, "")
	}
	//varDump(reflect.ValueOf(s), indent, "")
}

func varDump(value reflect.Value, indent string, preStr string) {
	var vKind = value.Kind()
	rIndent := indent
	if preStr != "" {
		rIndent = preStr
	}

	switch {
	case vKind == 0:
		fmt.Printf("%s[%s] : %v\n", rIndent, "nil", value)

	//bool + int +uint
	case vKind < 12:
		fmt.Printf("%s[%s] : %v\n", rIndent, value.Type(), value)

	//uintptr
	case vKind == 12:
		//fmt.Printf("%s[%s] ", r_indent, value.Type())
		varDump(value, indent, fmt.Sprintf("%s[%s] --> ", rIndent, value.Type()))

	//float + complex
	case vKind < 17:
		fmt.Printf("%s[%s] : %v\n", rIndent, value.Type(), value)

	//Array
	case vKind == 17:
		fmt.Printf("%sarray[%s](%d) ==>\n", rIndent, value.Type(), value.Len())
		index := 0
		for {
			if index >= value.Len() {
				break
			}
			//fmt.Printf("%s[%d] :", r_indent+"    ", index)
			varDump(value.Index(index), "    "+indent, fmt.Sprintf("%s[%d] :", rIndent+"    ", index))
			index++
		}

	//chan\func\Interface
	case vKind < 21:
		fmt.Printf("%s[%s]%v", rIndent, value.Type(), value)

	//Map
	case vKind == 21:
		fmt.Printf("%smap[%s](%d) ==>\n", rIndent, value.Type(), value.Len())
		keys := value.MapKeys()
		for _, key := range keys {
			//fmt.Printf("%s[%s] : ", indent+"    ", key)
			varDump(value.MapIndex(key), "    "+indent, fmt.Sprintf("%s\"%s\" : ", indent+"    ", key))
		}

	//ptr
	case vKind == 22:
		//fmt.Printf("%s[%s] -->\n", indent, value.Type())
		varDump(reflect.Indirect(value), indent, fmt.Sprintf("%s[%s] --> ", rIndent, value.Type()))

	//slice
	case vKind == 23:
		fmt.Printf("%sslice[%s](%d) ==>\n", rIndent, value.Type(), value.Len())
		//fmt.Printf("%s%s", indent+"    ", t.Field(k).Name)
		index := 0
		for {
			if index >= value.Len() {
				break
			}
			//fmt.Printf("%s[%d] : ", r_indent+"    ", index)
			varDump(value.Index(index), "    "+indent, fmt.Sprintf("%s[%d] : ", indent+"    ", index))
			index++
		}
		fmt.Printf("%s//end %s\n", indent, vKind.String())

	//string
	case vKind == 24:
		fmt.Printf("%s[%s](%d) ==> \"%s\"\n", rIndent, value.Type(), value.Len(), value)

	//Struct
	case vKind == 25:
		fmt.Printf("%sstruct[%s] ==> {\n", rIndent, value.Type())

		t := value.Type()
		for k := 0; k < value.NumField(); k++ {
			varDump(value.Field(k), indent+"    ", fmt.Sprintf("%s%s ", indent+"    ", t.Field(k).Name))
		}
		fmt.Printf("%s} //end %s\n", indent, vKind.String())

	//Unsafeptr
	case vKind == 26:
		//fmt.Printf("%s[%s] -->\n", r_indent, vKind.String())
		varDump(reflect.Indirect(value), indent, fmt.Sprintf("%s[%s] --> ", rIndent, vKind.String()))

	default:
		fmt.Printf("%s[%s] : %v\n", rIndent, "Unknown", value)
	}

	return
}

func GetFuncName() (string, error) {
	pc, _, _, succ := runtime.Caller(1)
	if !succ {
		return "", errors.New("get current function name failed")
	}

	return runtime.FuncForPC(pc).Name(), nil
}
