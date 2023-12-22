package email

import "strings"

func IsValid(email string) bool {
	// todo user REGEX to validation Email

	if !strings.Contains(email, "@") {
		return false
	}

	var splitEmail = strings.Split(email, "@")
	if len(splitEmail) != 2 {
		return false
	}

	if len(splitEmail[0]) <= 4 || len(splitEmail[1]) >= 50 || len(splitEmail[1]) <= 2 {
		return false
	}
	return true
}
