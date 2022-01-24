package frontend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var additionalPropertiesPattern = regexp.MustCompile(`additionalProperties '(.*)' not allowed`)

func jsonPathToYAML(path string) string {
	var format strings.Builder
	for _, elem := range strings.Split(path, "/")[1:] {
		num, err := strconv.ParseUint(elem, 10, 64)
		if err != nil {
			format.WriteString("." + elem)
		} else {
			format.WriteString(fmt.Sprintf("[%d]", num))
		}
	}
	return "$" + format.String()
}
