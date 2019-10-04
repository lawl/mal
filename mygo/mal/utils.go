package mal

import (
	"fmt"
	"strings"
)

// Stuff that doesn't fit anywhere else

//TypeToHashKey Takes a mal type, and if it's a string or keyword, returns the raw string for use in a hash map
func TypeToHashKey(key Type) (string, error) {
	if strKey, ok := key.(*String); ok {
		return strKey.Value, nil
	}
	if strKey, ok := key.(*Keyword); ok {
		return strKey.Value, nil
	}
	return "", fmt.Errorf("Map keys must be of type String, got '%T' instead", key)

}

//NativeStringToMalHashKey takes a string and returns either a malString or malKeyword
func NativeStringToMalHashKey(str string) Type {
	if strings.HasPrefix(str, ":") {
		return &Keyword{Value: str}
	}
	return &String{Value: str}
}
