package errutil

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMultiError_Normalize(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	err4 := errors.New("error 4")
	err5 := errors.New("error 5")

	errs := MultiError{
		nil, err1, err2, nil,
		MultiError{
			err3, nil,
			MultiError{
				nil, nil,
				MultiError{err4, err5},
			},
		},
	}

	assert.EqualValues(t, MultiError{err1, err2, err3, err4, err5}, errs.Normalize())
}

func TestCollectErrors(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	collected := CollectErrors(err1, err2, err3)
	assert.EqualValues(t, MultiError{err1, err2, err3}, collected)

	assert.Equal(t, err1, CollectErrors(err1, nil, nil))

	assert.Equal(t, MultiError{err1, err1, err2, err3}, CollectErrors(err1, collected))
}

func TestMultiError_Error(t *testing.T) {
	errs := CollectErrors(
		errors.New("something bad happened"),
		errors.New("another thing happened"),
		errors.New("terrible thing happened"),
	)

	assert.NotNil(t, errs)
	errStr := errs.Error()
	assert.Contains(t, errStr, "something bad happened")
	assert.Contains(t, errStr, "another thing happened")
	assert.Contains(t, errStr, "terrible thing happened")
}
