package errutils

import (
	"fmt"
	"strings"
)

// MultiError is an error which contains multiple errors within it.
type MultiError []error

// Normalize removes all instances of nil from the MultiError and flattens
// nested MultiError instances.
func (e MultiError) Normalize() MultiError {
	errs := make(MultiError, 0, len(e))

	for _, err := range e {
		if err == nil {
			continue
		}

		if merr, ok := err.(MultiError); ok {
			errs = append(errs, merr.Normalize()...)
		} else {
			errs = append(errs, err)
		}
	}

	return errs
}

// Strings returns a slice of all error strings in the MultiError.
// This does not use a normalized version of the MultiError.
func (e MultiError) Strings() []string {
	errStrings := make([]string, len(e))
	for i, err := range e {
		errStrings[i] = err.Error()
	}

	return errStrings
}

func (e MultiError) Error() string {
	errStrs := e.Strings()
	for i, s := range errStrs {
		errStrs[i] = fmt.Sprintf("  %s", s)
	}

	return fmt.Sprintf("Mulltiple errors:\n%s", strings.Join(errStrs, "\n"))
}

// AsError returns nil if the normalised MultiError is empty, the only error if
// there's exactly one error and otherwise it returns the normalised version of
// itself.
func (e MultiError) AsError() error {
	norm := e.Normalize()
	if len(norm) == 0 {
		return nil
	} else if len(norm) == 1 {
		return norm[0]
	} else {
		return norm
	}
}

// CollectErrors creates a MultiError from the given errors and returns its
// error representation (AsError).
// If there are no non-nil errors it returns nil, if there's exactly one error,
// it is returned and if there are more errors the MultiError is returned.
func CollectErrors(errs ...error) error {
	return MultiError(errs).AsError()
}
