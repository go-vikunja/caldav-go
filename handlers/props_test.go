package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/samedi/caldav-go/test"
)

const propsTestData = "BEGIN:VEVENT\nSUMMARY:Party\nEND:VEVENT"

func propsTestStorage() test.FakeStorage {
	stg := test.NewFakeStorage()
	stg.AddFakeResource("/test-data/props/", "123-456-789.ics", propsTestData)
	return stg
}

func TestPropfindAnswersRepeatedPropsOnce(t *testing.T) {
	handler := propfindHandler{
		handlerData{
			requestPath: "/test-data/props/123-456-789.ics",
			requestBody: `
			<?xml version="1.0" encoding="UTF-8"?>
			<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
				<D:prop>
					<C:calendar-data/>
					<D:getetag/>
					<C:calendar-data/>
					<C:calendar-data/>
				</D:prop>
			</D:propfind>
			`,
			response: NewResponse(),
			storage:  propsTestStorage(),
		},
	}

	resp := handler.Handle()
	test.AssertInt(resp.Status, 207, t)
	test.AssertInt(strings.Count(resp.Body, "<C:calendar-data>"), 1, t)
	test.AssertInt(strings.Count(resp.Body, "<D:getetag>"), 1, t)
}

func TestReportAnswersRepeatedPropsOnce(t *testing.T) {
	handler := reportHandler{
		handlerData{
			requestPath: "/test-data/props/",
			requestBody: `
			<?xml version="1.0" encoding="UTF-8"?>
			<C:calendar-multiget xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
				<D:prop>
					<C:calendar-data/>
					<C:calendar-data/>
				</D:prop>
				<D:prop>
					<C:calendar-data/>
				</D:prop>
				<D:href>/test-data/props/123-456-789.ics</D:href>
			</C:calendar-multiget>
			`,
			response: NewResponse(),
			storage:  propsTestStorage(),
		},
	}

	resp := handler.Handle()
	test.AssertInt(resp.Status, 207, t)
	test.AssertInt(strings.Count(resp.Body, "<C:calendar-data>"), 1, t)
}

func TestPropsWithTheSameNameInOtherNamespacesAreKept(t *testing.T) {
	handler := propfindHandler{
		handlerData{
			requestPath: "/test-data/props/123-456-789.ics",
			requestBody: `
			<?xml version="1.0" encoding="UTF-8"?>
			<D:propfind xmlns:D="DAV:" xmlns:X="urn:example">
				<D:prop>
					<D:getetag/>
					<X:getetag/>
				</D:prop>
			</D:propfind>
			`,
			response: NewResponse(),
			storage:  propsTestStorage(),
		},
	}

	resp := handler.Handle()
	test.AssertInt(resp.Status, 207, t)
	test.AssertInt(strings.Count(resp.Body, "<D:getetag>"), 1, t)
	test.AssertInt(strings.Count(resp.Body, "<getetag/>"), 1, t)
}

func manyProps(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "<X:p%d/>", i)
	}
	return b.String()
}

func TestTooManyRequestedProps(t *testing.T) {
	propfindBody := func(props string) string {
		return `<D:propfind xmlns:D="DAV:" xmlns:X="urn:example"><D:prop>` + props + `</D:prop></D:propfind>`
	}
	reportBody := func(props string) string {
		return `<C:calendar-multiget xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:X="urn:example"><D:prop>` +
			props + `</D:prop><D:href>/test-data/props/123-456-789.ics</D:href></C:calendar-multiget>`
	}

	handle := func(method, path, body string) *Response {
		data := handlerData{requestPath: path, requestBody: body, response: NewResponse(), storage: propsTestStorage()}
		if method == "PROPFIND" {
			return propfindHandler{data}.Handle()
		}
		return reportHandler{data}.Handle()
	}

	cases := []struct {
		method string
		path   string
		body   func(string) string
	}{
		{"PROPFIND", "/test-data/props/123-456-789.ics", propfindBody},
		{"REPORT", "/test-data/props/", reportBody},
	}

	for _, c := range cases {
		atLimit := handle(c.method, c.path, c.body(manyProps(MaxRequestedProps)))
		test.AssertInt(atLimit.Status, 207, t)

		// Repeats do not count against the limit.
		repeated := handle(c.method, c.path, c.body(manyProps(MaxRequestedProps)+manyProps(MaxRequestedProps)))
		test.AssertInt(repeated.Status, 207, t)

		overLimit := handle(c.method, c.path, c.body(manyProps(MaxRequestedProps+1)))
		test.AssertInt(overLimit.Status, http.StatusBadRequest, t)
	}
}

func TestReportAnswersRepeatedHrefsOnce(t *testing.T) {
	handler := reportHandler{
		handlerData{
			requestPath: "/test-data/props/",
			requestBody: `
			<?xml version="1.0" encoding="UTF-8"?>
			<C:calendar-multiget xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
				<D:prop>
					<C:calendar-data/>
				</D:prop>
				<D:href>/test-data/props/123-456-789.ics</D:href>
				<D:href>/test-data/props/123-456-789.ics</D:href>
				<D:href>/test-data/props/123-456-789.ics</D:href>
			</C:calendar-multiget>
			`,
			response: NewResponse(),
			storage:  propsTestStorage(),
		},
	}

	resp := handler.Handle()
	test.AssertInt(resp.Status, 207, t)
	test.AssertInt(strings.Count(resp.Body, "<C:calendar-data>"), 1, t)
}
