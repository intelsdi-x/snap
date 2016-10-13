// +build small

package core

import (
	"testing"

	"github.com/intelsdi-x/snap/pkg/stringutils"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricSeparator(t *testing.T) {
	tc := getTestCases()
	Convey("Test namespace separator", t, func() {
		for _, c := range tc {
			Convey("namespace "+c.input.String(), func() {
				firstChar := stringutils.GetFirstChar(c.input.String())
				So(firstChar, ShouldEqual, c.expected)
			})
		}
	})
}

type testCase struct {
	input    Namespace
	expected string
}

// getTestCases tests the namespace and nsPriorityList.
func getTestCases() []testCase {
	tcs := []testCase{
		testCase{
			input:    NewNamespace("/hello", "/world"),
			expected: "|",
		},
		testCase{
			input:    NewNamespace("/hello", "/world", "corporate-service|"),
			expected: "%",
		},
		testCase{
			input:    NewNamespace("/hello", "/world", "|corporate-service%", "monday_to_friday"),
			expected: ":",
		},
		testCase{
			input:    NewNamespace("/hello", "/world", "corporate-service/%|-_^><+=:;&", "monday_friday", "㊽ÄA小ヒ☍"),
			expected: "大",
		},
		testCase{
			input:    NewNamespace("A小ヒ☍小㊽%:;", "/hello", "/world大|", "monday_friday", "corporate-service"),
			expected: "^",
		},
	}
	return tcs
}
