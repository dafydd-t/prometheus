// Copyright 2020 The Prometheus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"bytes"
	"fmt"
	"io"
)

// multiError type allows to combine multiple errors into one.
type multiError []error

// NewMulti returns multiError with provided errors added if not nil.
func NewMulti(errs ...error) multiError { // nolint:golint
	m := multiError{}
	m.Add(errs...)
	return m
}

// Add adds the error to the error list if it is not nil.
// If the error has a nonNilMultiError error, errors from passed multi errors are combine in the
// same level as existing errors.
func (es *multiError) Add(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		if merr, ok := err.(nonNilMultiError); ok {
			*es = append(*es, merr.errs...)
			continue
		}
		*es = append(*es, err)
	}
}

// Err returns the error list as an error or nil if it is empty.
func (es multiError) Err() error {
	if len(es) == 0 {
		return nil
	}
	return nonNilMultiError{errs: es}
}

// nonNilMultiError implements the error interface, and it represents
// multiError with at least one error inside it.
// This type is needed to make sure that nil is returned when no error is combined in multiError for err != nil
// check to work.
type nonNilMultiError struct {
	errs multiError
}

// Error returns a concatenated string of the contained errors.
func (es nonNilMultiError) Error() string {
	var buf bytes.Buffer

	if len(es.errs) > 1 {
		fmt.Fprintf(&buf, "%d errors: ", len(es.errs))
	}

	for i, err := range es.errs {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}

	return buf.String()
}

// CloseAll closed all given closes with all errors recorded. All closers are closed.
func CloseAll(cs []io.Closer) error {
	errs := NewMulti()
	for _, c := range cs {
		errs.Add(c.Close())
	}
	return errs.Err()
}
