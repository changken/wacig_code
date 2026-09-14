package lexer

import "monkey/token"

type Lexer struct {
	input        string
	position     int  //目前讀取到的字元位置
	readPosition int  //下一個要讀取的字元位置
	ch           byte //目前讀取的字元
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	//如果下一個要讀取的字元位置大於等於輸入字串的長度，則將目前讀取的字元設為0
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		//否則
		//將目前讀取的字元設為輸入字串中下一個要讀取的字元
		l.ch = l.input[l.readPosition]
	}
	//目前的讀取位置設為下一個要讀取的字元位置
	l.position = l.readPosition
	//下一個要讀取的字元位置加1
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	//跳過空白字元
	l.skipWhitespace()

	//看目前字元是甚麼符號
	switch l.ch {
	case '=':
		//如果下一個字元是=，則回傳EQ
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		//如果下一個字元是=，則回傳NOT_EQ
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
		//數組
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	//哈希用
	case ':':
		tok = newToken(token.COLON, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		//如果是0，代表已經讀取到輸入字串的結尾
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		//如果目前讀取的字元是字母，則讀取識別符
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			//如果是識別符號，則回傳對應的TokenType
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
			//如果不是識別符號 就是ILLEGAL 非法符號
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{
		Type:    tokenType,
		Literal: string(ch),
	}
}

func (l *Lexer) readIdentifier() string {
	//取得目前的position
	position := l.position
	//如果目前讀取的字元是字母，則繼續讀取下一個字元
	for isLetter(l.ch) {
		l.readChar()
	}
	//把該區段的字元 應該說是substring
	//從目前的position到目前讀取的字元位置
	//回傳該區段的字元
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// 如果是空白字元，則繼續讀取下一個字元
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	//抓取目前的position
	position := l.position
	//如果是數字，則繼續讀取下一個字元
	for isDigit(l.ch) {
		l.readChar()
	}
	//回傳該區段的字元
	//從目前的position到目前讀取的字元位置
	return l.input[position:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// 窺視下一個字元，僅看下一個字元 不移動pointor
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

// 讀取字串
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		//讀取字元
		l.readChar()
		//直到遇到"
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	//從第一個"到第二個" index-1之間就是字串
	return l.input[position:l.position]
}
