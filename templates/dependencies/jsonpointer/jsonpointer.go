// Package jsonpointer provides the jsonpointer.5.0.1.js file.
// Generated from https://bundlejs.com using the following code
/*
import { get, set } from 'jsonpointer';

globalThis.jsonpointer = {
  get: get,
  set: set,
};
*/
package jsonpointer

import (
	_ "embed"
)

//nolint:revive
//go:embed jsonpointer.5.0.1.js
var JS string
