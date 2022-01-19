package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	androidRegexp = regexp.MustCompile("Android")
	MSIERegexp    = regexp.MustCompile("MSIE")
)

type User struct {
	Name     string
	Email    string
	Browsers []string
}

// optimized version of ShowSearch
func FastSearch(out io.Writer) {

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	// hash table for storing found browsers
	browsers := make(map[string]struct{})

	decoder := json.NewDecoder(file)

	fmt.Fprintf(out, "found users:\n")

	i := -1
	var user User
	for decoder.More() {
		i++

		// decode user
		err := decoder.Decode(&user)
		if err != nil {
			panic(err)
		}

		if len(user.Browsers) == 0 {
			continue
		}

		// analysis of user browsers
		var isAndroid, isMSIE bool
		for _, browser := range user.Browsers {
			if androidRegexp.MatchString(browser) {
				isAndroid = true
				browsers[browser] = struct{}{}
			}
			if MSIERegexp.MatchString(browser) {
				isMSIE = true
				browsers[browser] = struct{}{}
			}
		}

		// skip the user if he does not use Android and MSIE
		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(browsers))
}
