package errtemplates

import "github.com/pkg/errors"

const (
	requiredTagFailTemplate     = "Validation for '%s' failed: can not be blank"
	fileTagFailTemplate         = "Validation for '%s' failed: '%s' does not exist"
	urlTagFailTemplate          = "Validation for '%s' failed: '%s' incorrect url"
	oneofTagFailTemplate        = "Validation for '%s' failed: '%s' is not one of the options"
	hostnamePortTagFailTemplate = "Validation for '%s' failed: '%s' incorrect hostname and port"
)

func RequiredTagFail(cause string) error {
	return errors.Errorf(requiredTagFailTemplate, cause)
}

func FileTagFail(cause string, value interface{}) error {
	return errors.Errorf(fileTagFailTemplate, cause, value)
}

func UrlTagFail(cause string, value interface{}) error {
	return errors.Errorf(urlTagFailTemplate, cause, value)
}

func OneofTagFail(cause string, value interface{}) error {
	return errors.Errorf(oneofTagFailTemplate, cause, value)
}

func HostnamePortTagFail(cause string, value interface{}) error {
	return errors.Errorf(hostnamePortTagFailTemplate, cause, value)
}
