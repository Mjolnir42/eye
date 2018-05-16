/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package v2 // import "github.com/mjolnir42/eye/lib/eye.proto/v2"

import "time"

// PosTimeInf is the time which should be considered +infinity
//
// This package variable must be set prior to using any of the Format* or
// Parse* functions
var PosTimeInf time.Time

// NegTimeInf is the time which should be considered -infinity
//
// This package variable must be set prior to using any of the Format* or
// Parse* functions
var NegTimeInf time.Time

// TimeFormatString is the string that should be used to format
// timestamps via time.Time.Format
//
// This package variable must be set prior to using any of the Format* or
// Parse* functions
var TimeFormatString string

// FormatValidity returns t formatted with TimeFormatString, unless t is
// PosTimeInf (= "forever") or NegTimeInf (= "never")
func FormatValidity(t time.Time) string {
	if t.Equal(PosTimeInf) || t.After(PosTimeInf) {
		return `forever`
	}
	if t.Equal(NegTimeInf) || t.Before(NegTimeInf) {
		return `never`
	}

	return t.Format(TimeFormatString)
}

// ParseValidity parses a time string formatted by FormatValidity
func ParseValidity(t string) (z time.Time) {
	switch t {
	case `forever`:
		z, _ = time.Parse(TimeFormatString, PosTimeInf.Format(TimeFormatString))
	case `never`:
		z, _ = time.Parse(TimeFormatString, NegTimeInf.Format(TimeFormatString))
	default:
		z, _ = time.Parse(TimeFormatString, t)
	}
	return
}

// FormatProvision returns t formatted with TimeFormatString, unless t
// is PosTimeInf (= "never") or NegTimeInf (= "always")
func FormatProvision(t time.Time) string {
	if t.Equal(PosTimeInf) || t.After(PosTimeInf) {
		return `never`
	}
	if t.Equal(NegTimeInf) || t.Before(NegTimeInf) {
		return `always`
	}

	return t.Format(TimeFormatString)
}

// ParseProvision parses a time string formatted by FormatProvision
func ParseProvision(t string) (z time.Time) {
	switch t {
	case `never`:
		z, _ = time.Parse(TimeFormatString, PosTimeInf.Format(TimeFormatString))
	case `always`:
		z, _ = time.Parse(TimeFormatString, NegTimeInf.Format(TimeFormatString))
	default:
		z, _ = time.Parse(TimeFormatString, t)
	}
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
