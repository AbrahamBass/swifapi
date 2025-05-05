package testifyx

import (
	"fmt"
	"reflect"
	"strings"
)

type TestCase interface {
	Name() string
}

type ParamTestCase struct {
	name string
	Data interface{}
}

func (ptc *ParamTestCase) Name() string {
	return ptc.name
}

func (ts *TestSuite) WithCases(cases interface{}, testFunc interface{}) *TestSuite {
	ts.t.Helper()

	casesVal := reflect.ValueOf(cases)
	if casesVal.Kind() != reflect.Slice {
		panic("WithCases requires a slice of test cases")
	}

	testFuncVal := reflect.ValueOf(testFunc)
	if testFuncVal.Kind() != reflect.Func {
		panic("Test function required")
	}

	for i := 0; i < casesVal.Len(); i++ {
		caseData := casesVal.Index(i).Interface()
		caseName := ts.generateCaseName(caseData, i)

		ts.It(caseName, func(tc *TC) {
			params := []reflect.Value{
				reflect.ValueOf(tc),
				reflect.ValueOf(caseData),
			}
			testFuncVal.Call(params)
		})
	}
	return ts
}

func (ts *TestSuite) generateCaseName(caseData interface{}, index int) string {
	if nameable, ok := caseData.(TestCase); ok {
		return nameable.Name()
	}

	val := reflect.ValueOf(caseData)
	if val.Kind() == reflect.Struct {
		var parts []string
		for i := 0; i < val.NumField(); i++ {
			parts = append(parts, fmt.Sprintf("%v", val.Field(i).Interface()))
		}
		return fmt.Sprintf("Case %d: [%s]", index, strings.Join(parts, ", "))
	}

	return fmt.Sprintf("Case %d: %+v", index, caseData)
}
