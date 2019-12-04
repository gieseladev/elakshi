/*
Package service provides common service functionality.
*/
package service

// Identifier is an interface providing a service id.
type Identifier interface {
	// ServiceID returns the id of the service.
	ServiceID() string
}
