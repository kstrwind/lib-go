package kdebug

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

func VarDump(s interface{}) {
	indent := ""

	var_dump(reflect.ValueOf(s), indent, "")
}

func var_dump(value reflect.Value, indent string, prestr string) {
	var vKind = value.Kind()
	r_indent := indent
	if prestr != "" {
		r_indent = prestr
	}

	switch {
	case vKind == 0:
		fmt.Printf("%s[%s] : %v\n", r_indent, "nil", value)

	//bool + int +uint
	case vKind < 12:
		fmt.Printf("%s[%s] : %v\n", r_indent, value.Type(), value)

	//uintptr
	case vKind == 12:
		//fmt.Printf("%s[%s] ", r_indent, value.Type())
		var_dump(value, indent, fmt.Sprintf("%s[%s] --> ", indent, value.Type()))

	//float + complex
	case vKind < 17:
		fmt.Printf("%s[%s] : %v\n", r_indent, value.Type(), value)

	//Array
	case vKind == 17:
		fmt.Printf("%sarray[%s](%d) ==>\n", r_indent, value.Type(), value.Len())
		index := 0
		for {
			if index >= value.Len() {
				break
			}
			//fmt.Printf("%s[%d] :", r_indent+"    ", index)
			var_dump(value.Index(index), "    "+indent, fmt.Sprintf("%s[%d] :", r_indent+"    ", index))
			index++
		}

	//chan\func\Interface
	case vKind < 21:
		fmt.Printf("%s[%s]%v", r_indent, value.Type(), value)

	//Map
	case vKind == 21:
		fmt.Printf("%smap[%s](%d) ==>\n", r_indent, value.Type(), value.Len())
		keys := value.MapKeys()
		for _, key := range keys {
			//fmt.Printf("%s[%s] : ", indent+"    ", key)
			var_dump(value.MapIndex(key), "    "+indent, fmt.Sprintf("%s[%s] : ", indent+"    ", key))
		}

	//ptr
	case vKind == 22:
		//fmt.Printf("%s[%s] -->\n", indent, value.Type())
		var_dump(reflect.Indirect(value), indent, fmt.Sprintf("%s[%s] --> ", indent, value.Type()))

	//slice
	case vKind == 23:
		fmt.Printf("%sslice[%s](%d) ==>\n", indent, value.Type(), value.Len())
		//fmt.Printf("%s%s", indent+"    ", t.Field(k).Name)
		index := 0
		for {
			if index >= value.Len() {
				break
			}
			//fmt.Printf("%s[%d] : ", r_indent+"    ", index)
			var_dump(value.Index(index), "    "+indent, fmt.Sprintf("%s[%d] : ", indent+"    ", index))
			index++
		}
		fmt.Printf("%s//end %s\n", indent, vKind.String())

	//string
	case vKind == 24:
		fmt.Printf("%s[%s](%d) ==> \"%s\"\n", r_indent, value.Type(), value.Len(), value)

	//Struct
	case vKind == 25:
		fmt.Printf("%sstruct[%s] ==> {\n", r_indent, value.Type())
		t := value.Type()
		for k := 0; k < t.NumField(); k++ {
			var_dump(value.Field(k), indent+"    ", fmt.Sprintf("%s%s ", indent+"    ", t.Field(k).Name))
		}
		fmt.Printf("%s} //end %s\n", indent, vKind.String())

	//Unsafeptr
	case vKind == 26:
		//fmt.Printf("%s[%s] -->\n", r_indent, vKind.String())
		var_dump(reflect.Indirect(value), indent, fmt.Sprintf("%s[%s] --> ", indent, vKind.String()))

	default:
		fmt.Printf("%s[%s] : %v\n", r_indent, "Unknown", value)
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
