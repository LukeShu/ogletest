// Copyright 2011 Aaron Jacobs. All Rights Reserved.
// Author: aaronjjacobs@gmail.com (Aaron Jacobs)
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

package ogletest

import (
	"flag"
	"fmt"
	"github.com/jacobsa/ogletest/internal"
	"path"
	"reflect"
	"runtime"
	"testing"
)

var testFilter = flag.String("ogletest.run", "", "Regexp for matching tests to run.")

// runTest runs a single test, returning a slice of failure records for that test.
func runTest(suite interface{}, method reflect.Method) (failures []internal.FailureRecord) {
	suiteValue := reflect.ValueOf(suite)
	suiteType := suiteValue.Type()
	suiteName := suiteType.Elem().Name()

	fmt.Printf("==== %s.%s\n", suiteName, method.Name)

	// Set up a clean slate for this test.
	internal.CurrentTest = internal.NewTestState()

	defer func() {
		// Return the failures the test recorded, whether it panics or not. If it
		// panics, additionally return a failure for the panic.
		failures = internal.CurrentTest.FailureRecords
		if r := recover(); r != nil {
			// The stack looks like this:
			//
			//     <this deferred function>
			//     panic(r)
			//     <function that called panic>
			//
			_, fileName, lineNumber, ok := runtime.Caller(2)
			var panicRecord internal.FailureRecord
			if ok {
				panicRecord.FileName = path.Base(fileName)
				panicRecord.LineNumber = lineNumber
			}

			panicRecord.GeneratedError = fmt.Sprintf("panic: %v", r)
			failures = append(failures, panicRecord)
		}

		// Reset the global CurrentTest state, so we don't accidentally use it elsewhere.
		internal.CurrentTest = nil
	}()

	// Create a receiver, and call it.
	suiteInstance := reflect.New(suiteType.Elem())
	runMethodIfExists(suiteInstance, "SetUp")
	runMethodIfExists(suiteInstance, method.Name)
	runMethodIfExists(suiteInstance, "TearDown")

	// The return value is set in the deferred function above.
	return
}

// RunTests runs the test suites registered with ogletest, communicating
// failures to the supplied testing.T object. This is the bridge between
// ogletest and the testing package (and gotest); you should ensure that it's
// called at least once by creating a gotest-compatible test function and
// calling it there.
//
// For example:
//
//     import (
//       "github.com/jacobsa/ogletest"
//       "testing"
//     )
//
//     func TestOgletest(t *testing.T) {
//       ogletest.RunTests(t)
//     }
//
func RunTests(t *testing.T) {
	for _, suite := range testSuites {
		val := reflect.ValueOf(suite)
		typ := val.Type()
		suiteName := typ.Elem().Name()

		fmt.Println("=========", suiteName)

		// Run the SetUpTestSuite method, if any.
		runMethodIfExists(val, "SetUpTestSuite")

		// Run each method.
		//
		// TODO(jacobsa): Recover from panics.
		// TODO(jacobsa): Pay attention to failures.
		// TODO(jacobsa): Confirm that unexported methods don't show up here.
		for i := 0; i < typ.NumMethod(); i++ {
			method := typ.Method(i)
			if isSpecialMethod(method.Name) {
				continue
			}

			// Run the test.
			failures := runTest(suite, method)

			// Print any failures, and mark the test as having failed if there are any.
			for _, record := range failures {
				t.Fail()
				fmt.Printf(
					"%s:%d:\n%s\n%s",
					record.FileName,
					record.LineNumber,
					record.GeneratedError,
					record.UserError)
			}
		}

		// Run the TearDownTestSuite method, if any.
		runMethodIfExists(val, "TearDownTestSuite")
	}
}

func runMethodIfExists(v reflect.Value, name string) {
	method := v.MethodByName(name)
	if method.Kind() == reflect.Invalid {
		return
	}

	// TODO(jacobsa): Panic (or print error?) if method doesn't have the right
	// signature.
	method.Call([]reflect.Value{})
}

func isSpecialMethod(name string) bool {
	return (name == "SetUpTestSuite") ||
		(name == "TearDownTestSuite") ||
		(name == "SetUp") ||
		(name == "TearDown")
}
