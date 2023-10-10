package util

import "strings"

// AddValueBeforeExtension finds the extension and inserts ".token" before it.  NOTE that the
// token argument should not contain the preceding ".". This is added by the function.
//
// If the file has no exension, the token is just appended.
func AddValueBeforeExtension(filename string, token string) string {
	//set output file from input file
	tokens := strings.Split(filename, ".")
	if len(tokens) == 1 {
		//no extensions found, so append
		return filename + "." + token
	} else {
		//insert the new token at the second-to-last place and rejoin the tokens
		tokens = append(tokens, tokens[len(tokens)-1]) //add the last token to the end again
		tokens[len(tokens)-2] = token                  //set the old last token to the sealed token
		return strings.Join(tokens, ".")
	}

}
