package util

import (
	"net/url"
	"regexp"
)

//=================================================================================================
// Validation helper functions

// source: https://stackoverflow.com/questions/106179/regular-expression-to-match-dns-hostname-or-ip-address
var HostnameRe = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

// IsURLHost returns true if the candidate string is a valid hostname
func IsUrlHost(candidate string) bool {
	//we don't want the candidate to have to have a scheme, so we add our own
	candidateWithScheme := "varanus://" + candidate

	u, err := url.ParseRequestURI(candidateWithScheme)

	// fmt.Printf("%s --> %#v\n\n", candidate, u)

	//not a valid URL if:
	// - err is not nil
	// - the scheme is not the one we added
	// - the host is not the whole candidate string (this saves from having to check a bunch of
	//	 path variables in the u result)
	// - the candidate has anything other than letters, numbers, dashes, and dots

	if !HostnameRe.Match([]byte(candidate)) {
		return false
	}

	return err == nil && u.Scheme == "varanus" && u.Host == candidate
}
