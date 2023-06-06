package global

import "regexp"

var ValidType = regexp.MustCompile(`/update/(gauge|counter)/`)
var ValidID = regexp.MustCompile(`/update/.+/([a-zA-Z]*)/.*`)
var ValidValue = regexp.MustCompile(`/update/.+/.+/([+-]?[0-9]*[.]?[0-9]+)`)
