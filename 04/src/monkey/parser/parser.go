package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// 遞增的enum 1~7
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         //+
	PRODUCT     //*
	PREFIX      //-X or !X
	CALL        //myFunction(x)
	INDEX       // array[index]
)

// 優先次序 map
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,  //( 代表call函數有最高級別
	token.LBRACKET: INDEX, //代表 array[index] 最高級別
}

// Parser class
type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	//加入一個解析函數
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerIteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	//hasnmap
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)

	//加入中序解析
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	//index expression
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// 解析identifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// 注意 雙指針
// 等同於把目前的token用下一個取代
// 下一個peekToken 用parser 裡面的lexer的NextToken取代
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// 架構AST的根結點
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{} //List<>

	//如果current Token 不等於EOF
	//for p.curToken.Type != token.EOF {
	for !p.curTokenIs(token.EOF) {
		//解析statement
		stmt := p.parseStatement()
		//如果stmt不等於nil
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		//往下一個token
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//如果是let statement
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// 解析let statement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	//如果不是identitfer
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	//如果不為assign
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	//這邊要推進一個token 不然=會無法解析
	p.nextToken()

	//先不解吸expression, 如果還沒遇到;再繼續取下一個token
	// TODO: We're skipping the expressions until we
	//之前跳過let 後面的expression, 現在補回來
	stmt.Value = p.parseExpression(LOWEST)

	// 分號是可選的，有的話就往下推進一個token
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	//這邊先偷懶，先取出return -> ;之間的token 先讓他知道是return statement
	// TODO: We're skipping the expressions until we
	//之前省略了 現在補回來要解析他們的expression
	stmt.ReturnValue = p.parseExpression(LOWEST)

	// 分號是可選的，有的話就往下推進一個token
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	//下一個字元是不是分號 是的話在往下
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found.", t)

	p.errors = append(p.errors, msg)
}

// 檢查是有否類似的解析函數，如果是nil 回復nil
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	//如果有找到prefix 把她丟給leftExp
	leftExp := prefix()

	//如果下一個不是; 且 優先級別小於下一個token的優先級別
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		//中序去查表
		infix := p.infixParseFns[p.peekToken.Type]
		//中序等於nil
		if infix == nil {
			//回傳leftExpression
			return leftExp
		}

		//往下token
		p.nextToken()
		//把leftExp 丟給中序
		leftExp = infix(leftExp)
	}

	return leftExp
}

// 解析整數
func (p *Parser) parseIntegerIteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))

	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))

	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// 中序Expression
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))

	//先把當下的token放到infix
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	//優先級別
	precedence := p.curPrecedence()
	//往下個走
	p.nextToken()
	//右邊要用ParseExpression
	/*if expression.Operator == "+" {
		expression.Right = p.parseExpression(precedence - 1)
	} else {
		expression.Right = p.parseExpression(precedence)
	}*/
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	//如果下一個token不等於)
	if !p.expectPeek(token.RPAREN) {
		//回傳nil
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	//如果不為 (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST)

	//如果不為 )
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	//如果不為 {
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	//有沒有else
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		//開頭不為 {
		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	//不等於 } 且 不等於 EOF
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()

		//如果裡面的statement可以解析 則append
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		//往下token移動
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	//如果不是 (
	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	//解析parameters
	lit.Parameters = p.parseFunctionParameters()

	//如果不是 {
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	//解析body
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	//如果是 ) 回傳identifier list
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	//第一個 identifier
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	//如果是 , 往下2個token走
	//等於 ,identifier 一對
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	//如果下一個不等於 )
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	//exp.Arguments = p.parseCallArguments()
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	//如果是 ) 回傳整個args list
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	//第一個arguments
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	//第二個arguments
	//,expression 型式一阻
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	//如果找完不等於 )
	//這邊要注意 找到如果是 ) 就要推進token 沒有的話則nil
	if !p.expectPeek(end) {
		return nil
	}

	return list
}

/*func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	//如果是 ) 回傳整個args list
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	//第一個arguments
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	//第二個arguments
	//,expression 型式一阻
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	//如果找完不等於 )
	//這邊要注意 找到如果是 ) 就要推進token 沒有的話則nil
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}*/

// 把string token包裝為ast
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// 解析array
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}

	array.Elements = p.parseExpressionList(token.RBRACKET)

	return array
}

// 解析index expression
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	//遇到 [ 建立index expression
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	//往下一個token走
	p.nextToken()

	//右邊index處就拿去解析 expression patt演算法
	exp.Index = p.parseExpression(LOWEST)

	//如果不是 ] 則回傳nil
	//是的話就繼續推下一個token
	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	//如果遇到 {
	for !p.peekTokenIs(token.RBRACE) {
		//走下一個
		p.nextToken()
		//解析key expression
		key := p.parseExpression(LOWEST)

		//如果沒遇到: 則錯誤
		if !p.expectPeek(token.COLON) {
			return nil
		}

		//有遇到: 則往value去解析expression
		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		//如果下一個token不是} 且下一個token不是,(是,則推進token)
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	//如果解析完成沒遇到 }，則nil
	//有遇到則往下一個token走
	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	//解析完畢
	return hash
}

// 檢查當前的token type
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 檢查下一個token type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// 如果下一個token type是t
// , 他就true, 往下一個token走
// 否則false, 不往下一個token走
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		//如果預期的token不對 就要蒐集到error裡面
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s insteads.",
		t, p.peekToken.Type)

	p.errors = append(p.errors, msg)
}

// 看下一個token優先次序
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// 看目前token的優先次序
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
