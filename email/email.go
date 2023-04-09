package email

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type EmailAddress struct {
	User   string
	Domain string
	TLD    string
}

var (
	// rfc5322 is a RFC 5322 regex, as per: https://stackoverflow.com/a/201378/5405453.
	rfc5322 = regexp.MustCompile(
		fmt.Sprintf(
			"^%s*$",
			"(?i)(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])",
		),
	)
)

func Parse(email string) (*EmailAddress, error) {
	if !rfc5322.MatchString(email) {
		return nil, fmt.Errorf("format is incorrect for %s", email)
	}

	i := strings.LastIndexByte(email, '@')

	domain, tld, err := parseDomain(email[i+1:])
	if err != nil {
		return nil, err
	}

	e := &EmailAddress{
		User:   email[:i],
		Domain: domain,
		TLD:    tld,
	}
	return e, nil
}

func parseDomain(s string) (string, string, error) {
	url, err := url.Parse("http://" + s)
	if err != nil {
		return "", "", err
	}
	if url.Host == "" {
		return "", "", errors.New("blank host")
	}

	i := strings.Index(url.Host, ".")

	return url.Host[0:i], url.Host[i+1:], nil
}
