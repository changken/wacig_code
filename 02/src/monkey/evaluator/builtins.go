package evaluator

import (
	"fmt"
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	//內置len函數
	"len": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			//取出第一個arg參數
			switch arg := args[0].(type) {
			//添加array
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.String:
				//如果是string則回傳長度
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				//其他型態不支援
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	//內置first
	"first": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			//他的type要是array
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	//內置last
	"last": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			//他的type要是array
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				//取出最後一個
				return arr.Elements[length-1]
			}

			return NULL
		},
	},
	//內置rest
	"rest": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			//他的type要是array
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			if length > 0 {
				//建立一個新array
				newElements := make([]object.Object, length-1, length-1)
				//只copy 第一個到後面
				copy(newElements, arr.Elements[1:length])
				return &object.Array{Elements: newElements}
			}

			return NULL
		},
	},
	//內置push
	"push": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}

			//他的type要是array
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			//建立一個新array
			newElements := make([]object.Object, length+1, length+1)
			//copy一個新的array
			copy(newElements, arr.Elements)
			//把新的元素推到新copy array的最後一個
			newElements[length] = args[1]
			//回傳
			return &object.Array{Elements: newElements}
		},
	},
	//內置puts
	"puts": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			//分行印出
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			//回傳NULL
			return NULL
		},
	},
}
