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
	"sync"

	"github.com/jacobsa/oglemock"
)

// TestInfo represents information about a currently running or previously-run
// test.
type TestInfo struct {
	// A mock controller that is set up to report errors to the ogletest test
	// runner. This can be used for setting up mock expectations and handling
	// mock calls. The Finish method should not be run by the user; ogletest will
	// do that automatically after the test's TearDown method is run.
	//
	// Note that this feature is still experimental, and is subject to change.
	MockController oglemock.Controller

	// A mutex protecting shared state.
	mutex sync.RWMutex

	// A set of failure records that the test has produced.
	failureRecords []*failureRecord // Protected by mutex
}

// currentlyRunningTest is the state for the currently running test, if any.
var currentlyRunningTest *TestInfo

// newTestInfo creates a valid but empty TestInfo struct.
func newTestInfo() *TestInfo {
	info := &TestInfo{}
	info.failureRecords = make([]*failureRecord, 0)
	info.MockController = oglemock.NewController(&testInfoErrorReporter{info})
	return info
}

// testInfoErrorReporter is an oglemock.ErrorReporter that writes failure
// records into a test info struct.
type testInfoErrorReporter struct {
	testInfo *TestInfo
}

func (r *testInfoErrorReporter) ReportError(
	fileName string,
	lineNumber int,
	err error) {
	r.testInfo.mutex.Lock()
	defer r.testInfo.mutex.Unlock()

	record := &failureRecord{
		FileName:       fileName,
		LineNumber:     lineNumber,
		GeneratedError: err.Error(),
	}

	r.testInfo.failureRecords = append(r.testInfo.failureRecords, record)
}

func (r *testInfoErrorReporter) ReportFatalError(
	fileName string,
	lineNumber int,
	err error) {
	r.ReportError(fileName, lineNumber, err)
	panic(&assertThatError{})
}
