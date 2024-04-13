package lexer

// Convert string to a list of tokens.
func Tokenize(input string) ([]Token, error) {
	var tokens []Token
	lx := newLexer(input)
	for {
		t, err := lx.nextToken()
		switch err {
		case nil:
			tokens = append(tokens, t)
		case EoF{}:
			return tokens, nil
		default:
			return nil, err
		}
	}
}
