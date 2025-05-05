package testifyx

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/xeipuuv/gojsonschema"
)

func (tc *TC) AssertEqual(expected, actual interface{}) *TC {
	tc.t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		tc.t.Errorf("\n❌ Values not equal\nExpected: %+v\nReceived: %+v", expected, actual)
	}
	return tc
}

func (tc *TC) AssertTrue(condition bool, msg string) *TC {
	tc.t.Helper()
	if !condition {
		tc.t.Errorf("❌ AssertTrue failed: %s", msg)
	}
	return tc
}

func (tc *TC) AssertNil(obj interface{}) *TC {
	tc.t.Helper()
	if obj != nil {
		tc.t.Errorf("❌ Expected nil, got: %+v", obj)
	}
	return tc
}

func (tc *TC) AssertError(err error, expectedError string) *TC {
	tc.t.Helper()
	if err == nil {
		tc.t.Error("❌ Expected an error")
	} else if err.Error() != expectedError {
		tc.t.Errorf("\n❌ Incorrect error message\nExpected: %s\nReceived: %s",
			expectedError, err.Error())
	}
	return tc
}

/* ========== HTTP-SPECIFIC ASSERTIONS ========== */
func (hc *HC) AssertStatus(expected int) *HC {
	hc.t.Helper()
	if hc.response.StatusCode != expected {
		hc.t.Errorf("\n❌ Incorrect status code\nExpected: %d\nReceived: %d", expected, hc.response.StatusCode)
	}
	return hc
}

func (hc *HC) AssertJSON(expected interface{}) *HC {
	hc.t.Helper()

	var current, expectedMap map[string]interface{}

	if err := json.Unmarshal(hc.response.BodyBytes, &current); err != nil {
		hc.t.Errorf("❌ Failed to parse response JSON: %v", err)
		return hc
	}

	expectedBytes, _ := json.Marshal(expected)
	if err := json.Unmarshal(expectedBytes, &expectedMap); err != nil {
		hc.t.Errorf("❌ Invalid expected JSON: %v", err)
		return hc
	}

	if !reflect.DeepEqual(current, expectedMap) {
		hc.t.Errorf("\n❌ JSON mismatch\nExpected: %+v\nReceived: %+v", expectedMap, current)
	}

	return hc
}

func (hc *HC) AssertHeader(key, expected string) *HC {
	hc.t.Helper()

	current := hc.response.Header.Get(key)
	if current != expected {
		hc.t.Errorf("\n❌ Incorrect header %s\nExpected: %s\nReceived: %s", key, expected, current)
	}
	return hc
}

func (hc *HC) AssertBytes(expected []byte) *HC {
	hc.t.Helper()
	if !bytes.Equal(hc.response.BodyBytes, expected) {
		hc.t.Errorf("\n❌ Bytes mismatch\nExpected: %v\nReceived: %v", expected, hc.response.BodyBytes)
	}
	return hc
}

func (hc *HC) AssertInt(expected int) *HC {
	hc.t.Helper()
	body := string(hc.response.BodyBytes)
	actual, err := strconv.Atoi(body)
	if err != nil {
		hc.t.Errorf("\n❌ Could not convert response to integer: %v", err)
		return hc
	}
	if actual != expected {
		hc.t.Errorf("\n❌ Integer mismatch\nExpected: %d\nReceived: %d", expected, actual)
	}
	return hc
}

func (hc *HC) AssertString(expected string) *HC {
	hc.t.Helper()
	body := string(hc.response.BodyBytes)
	if body != expected {
		hc.t.Errorf("\n❌ String mismatch\nExpected: %q\nReceived: %q", expected, body)
	}
	return hc
}

func (hc *HC) AssertJSONSchema(schema map[string]interface{}) *HC {
	hc.t.Helper()

	schemaLoader := gojsonschema.NewGoLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(hc.response.BodyBytes)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		hc.t.Errorf("❌ Schema validation error: %v", err)
		return hc
	}

	if !result.Valid() {
		hc.t.Error("❌ JSON doesn't match schema:")
		for _, desc := range result.Errors() {
			hc.t.Errorf("  - %s", desc)
		}
	}

	return hc
}

func (hc *HC) AssertContentType(expected string) *HC {
	hc.t.Helper()
	contentType := hc.response.Header.Get("Content-Type")
	mimeType := strings.TrimSpace(strings.Split(contentType, ";")[0])

	if mimeType != expected {
		hc.t.Errorf("\n❌ Incorrect Content-Type\nExpected: %s\nReceived: %s",
			expected, mimeType)
	}
	return hc
}

func (hc *HC) AssertFileHeader(expectedType string) *HC {
	hc.t.Helper()

	fileSignatures := map[string][]byte{
		"PNG":  {0x89, 0x50, 0x4E, 0x47},
		"PDF":  []byte("%PDF-"),
		"JPEG": {0xFF, 0xD8, 0xFF},
		"ZIP":  {0x50, 0x4B, 0x03, 0x04},
	}

	expected, ok := fileSignatures[strings.ToUpper(expectedType)]
	if !ok {
		hc.t.Errorf("❌ Unsupported file type: %s", expectedType)
		return hc
	}

	if len(hc.response.BodyBytes) < len(expected) {
		hc.t.Errorf("❌ File too small for header validation")
		return hc
	}

	actualHeader := hc.response.BodyBytes[:len(expected)]
	if !bytes.Equal(actualHeader, expected) {
		hc.t.Errorf("\n❌ Invalid file header\nExpected: %X\nReceived: %X",
			expected, actualHeader)
	}

	return hc
}

func (hc *HC) AssertCookie(name, expectedValue string) *HC {
	hc.t.Helper()
	for _, cookie := range hc.response.Cookies() {
		if cookie.Name == name {
			if cookie.Value != expectedValue {
				hc.t.Errorf("\n❌ Cookie %s\nExpected: %s\nReceived: %s",
					name, expectedValue, cookie.Value)
			}
			return hc
		}
	}
	hc.t.Errorf("❌ Cookie not found: %s", name)
	return hc
}

func (hc *HC) AssertSecureCookie(name string) *HC {
	hc.t.Helper()
	for _, cookie := range hc.response.Cookies() {
		if cookie.Name == name {
			if !cookie.Secure || !cookie.HttpOnly {
				hc.t.Errorf("❌ Cookie %s not secure (Secure: %t, HttpOnly: %t)",
					name, cookie.Secure, cookie.HttpOnly)
			}
			return hc
		}
	}
	hc.t.Errorf("❌ Cookie not found: %s", name)
	return hc
}

func (hc *HC) AssertRedirectsTo(expectedURL string) *HC {
	if hc.response.StatusCode < 300 || hc.response.StatusCode >= 400 {
		hc.t.Errorf("❌ Not a redirect. Status code: %d", hc.response.StatusCode)
	}
	location, err := hc.response.Location()
	if err != nil {
		hc.t.Errorf("❌ Missing or invalid Location header")
	} else if location.String() != expectedURL {
		hc.t.Errorf("\n❌ Incorrect redirect\nExpected: %s\nReceived: %s",
			expectedURL, location.String())
	}
	return hc
}

func (hc *HC) AssertHTMLContains(selector, expectedContent string) *HC {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(hc.response.BodyBytes))
	if err != nil {
		hc.t.Errorf("❌ Error parsing HTML: %v", err)
		return hc
	}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if !strings.Contains(s.Text(), expectedContent) {
			hc.t.Errorf("\n❌ Missing content in %s\nExpected: %s",
				selector, expectedContent)
		}
	})

	return hc
}
