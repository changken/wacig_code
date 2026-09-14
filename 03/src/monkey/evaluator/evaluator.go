package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	//Statement
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.ReturnStatement:
		//先把return value本身拿去解析
		val := Eval(node.ReturnValue, env)
		//檢查是否回傳error
		if isError(val) {
			return val
		}
		//在包裝一層
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		//檢查是否有error
		if isError(val) {
			return val
		}
		//把該變數設為該值
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)

	//Expression
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		//檢查是否回傳error
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		//檢查是否回傳error
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		//檢查是否回傳error
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	//把ast包裝成object.Function
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		//如果有錯誤 回傳error
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	//把 string ast 包裝成 object string
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
		//數組
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		//如果只回傳一個 + 那個剛好是error
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		//把ast包裝成object
		return &object.Array{Elements: elements}
	case *ast.IndexExpression:
		//左邊拿去解析
		left := Eval(node.Left, env)
		//有錯誤直接返回
		if isError(left) {
			return left
		}
		//index拿去解析
		index := Eval(node.Index, env)
		//有錯誤也返回
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	//hashmap
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		/*if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}*/

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		//回傳error
		case *object.Error:
			return result
		}

	}

	return result
}

// 廢止
/*func evalStatement(stmts []ast.Statement) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement)

		//return value
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}

	return result
}*/

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			//如果回傳有returnvalue || error
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOpeatorExpression(right)
	default:
		//如果前置符號不為! - 的話 就是不知道的operator
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	//其實就是看!後面接什麼，接的把它變號
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOpeatorExpression(right object.Object) object.Object {
	//如果右邊的型態不為 integer, 前置-號會報錯誤
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s",
			right.Type())
	}

	//如果不為整數 則回傳null
	if right.Type() != object.INTEGER_OBJ {
		return NULL
	}

	//給一個整數object 變號
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		//如果operator不為上面的條件 則是不知道的operator
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	//字串中敘 (必須排在下面的 == / != pointer比較之前，否則字串永遠不會走到這裡)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	//如果中置兩者型態不一樣的話 應該要回傳type mismatch
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(
	ie *ast.IfExpression,
	env *object.Environment,
) object.Object {
	condition := Eval(ie.Condition, env)

	//檢查是否回傳error
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// 評估ident時 要把她取出來
func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {

	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", node.Value)
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		//有錯誤回傳錯誤
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		//蒐集argument
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		//新增一個函數的env
		extendedEnv := extendFunctionEnv(fn, args)
		//解析
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		//內置函數
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}

}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}

	return env
}
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	//如果不等於+
	if operator != "+" && operator != "==" && operator != "!=" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	//轉字串
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	//相加再回傳
	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	//注意這邊boolean是共用變數的狀態 要記得用helper
	case "==": //等於
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=": //不等於
		return nativeBoolToBooleanObject(leftVal != rightVal)
	}

	//錯誤 default => 自己補
	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	//如果left為array, index為integer
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	//hashmap + index expression
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		//如果左邊+index型態不對
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	//array本身
	arrayObject := array.(*object.Array)
	//index 必為 int
	idx := index.(*object.Integer).Value
	//最大數量
	max := int64(len(arrayObject.Elements) - 1)

	//如果idx 值域不在 [0, max] 之間
	if idx < 0 || idx > max {
		return NULL
	}

	//回傳該array的item
	return arrayObject.Elements[idx]
}

func evalHashLiteral(
	node *ast.HashLiteral,
	env *object.Environment,
) object.Object {
	//建立hashkey + hashpair
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		//key解析
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		//key是不是hashable
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		//value也拿去解析
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		//最後取出hashkey
		hashed := hashKey.HashKey()
		//把key-value塞進map
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	//先解析key
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	//在解析key value pair
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// 檢查obj是不是錯誤
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}

	return false
}
