package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUsername = regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString
	isValidFullname = regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString
)

func ValidateString(value string, minLength, maxLength int) error {
	if len(value) < minLength || len(value) > maxLength {
		return fmt.Errorf("string length must be between %d and %d characters", minLength, maxLength)
	}
	return nil
}

func ValidateUsername(value string) error {
	if err := ValidateString(value, 3, 32); err != nil {
		return err
	}

	if !isValidUsername(value) {
		return fmt.Errorf("username can only contain alphanumeric characters and underscores")
	}

	return nil
}

func ValidateFullname(value string) error {
	if err := ValidateString(value, 3, 32); err != nil {
		return err
	}

	if !isValidFullname(value) {
		return fmt.Errorf("fullname can only contain alphabetic characters and spaces")
	}

	return nil
}

func ValidatePassword(value string) error {
	if err := ValidateString(value, 6, 72); err != nil {
		return err
	}

	return nil
}

func ValidateEmail(value string) error {
	if err := ValidateString(value, 6, 128); err != nil {
		return err
	}

	_, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("invalid email format: %s", err)
	}

	return nil
}

