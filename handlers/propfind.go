package handlers

import (
	"encoding/xml"
	"net/http"
)

type propfindHandler struct {
	handlerData
}

func (ph propfindHandler) Handle() *Response {
	// read body string to xml struct
	type XMLProp2 struct {
		Tags []xml.Name `xml:",any"`
	}
	type XMLRoot2 struct {
		XMLName xml.Name
		Prop    XMLProp2 `xml:"DAV: prop"`
	}
	var requestXML XMLRoot2
	xml.Unmarshal([]byte(ph.requestBody), &requestXML)

	props, ok := uniqueProps(requestXML.Prop.Tags)
	if !ok {
		return ph.response.Set(http.StatusBadRequest, "")
	}

	// get the target resources based on the request URL
	resources, err := ph.storage.GetResources(ph.requestPath, ph.headers.IsDeep())
	if err != nil {
		return ph.response.SetError(err)
	}

	multistatus := &multistatusResp{
		Minimal:             ph.headers.IsMinimal(),
		user:                ph.user,
		supportedComponents: ph.supportedComponents,
	}
	// for each href, build the multistatus responses
	for _, resource := range resources {
		propstats := multistatus.Propstats(&resource, props)
		multistatus.AddResponse(resource.Path, true, propstats)
	}

	if multistatus.Minimal {
		ph.response.SetHeader(HD_PREFERENCE_APPLIED, HD_PREFER_MINIMAL)
	}

	return ph.response.Set(207, multistatus.ToXML())
}
