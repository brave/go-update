package v3

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/brave/go-update/extension"
)

// Request represents an Omaha v3 update request
type Request []extension.Extension

// UnmarshalJSON decodes the update server request JSON data
func (r *Request) UnmarshalJSON(b []byte, protocolVersion string) error {
	type Package struct {
		FP string `json:"fp"`
	}
	type Packages struct {
		Package []Package `json:"package"`
	}
	type App struct {
		AppID    string   `json:"appid"`
		FP       string   `json:"fp"`
		Version  string   `json:"version"`
		Packages Packages `json:"packages"`
	}
	type RequestWrapper struct {
		OS       string `json:"@os"`
		Updater  string `json:"@updater"`
		App      []App  `json:"app"`
		Protocol string `json:"protocol"`
	}
	type JSONRequest struct {
		Request RequestWrapper `json:"request"`
	}

	request := JSONRequest{}
	err := json.Unmarshal(b, &request)
	if err != nil {
		return err
	}

	*r = Request{}
	for _, app := range request.Request.App {
		fp := app.FP
		// spec discrepancy: FP might be set within a "package" object (v3) instead of the "app" object (v3.1)
		// https://github.com/google/omaha/blob/main/doc/ServerProtocolV3.md#package-request
		// https://chromium.googlesource.com/chromium/src.git/+/master/docs/updater/protocol_3_1.md#update-checks-body-update-check-request-objects-update-check-request-3
		if fp == "" && len(app.Packages.Package) > 0 {
			fp = app.Packages.Package[0].FP
		}
		*r = append(*r, extension.Extension{
			ID:      app.AppID,
			FP:      fp,
			Version: app.Version,
		})
	}

	// Verify the protocol version
	if request.Request.Protocol != protocolVersion {
		err = fmt.Errorf("request version: %v not supported by %v handler", request.Request.Protocol, protocolVersion)
	}

	return err
}

// UnmarshalXML decodes the update server request XML data
func (r *Request) UnmarshalXML(d *xml.Decoder, start xml.StartElement, protocolVersion string) error {
	// Common XML elements
	type UpdateCheck struct {
		XMLName xml.Name `xml:"updatecheck"`
	}

	// Version-specific types
	var apps []extension.Extension

	if protocolVersion == "3.0" {
		type Package struct {
			XMLName xml.Name `xml:"package"`
			FP      string   `xml:"fp,attr"`
		}
		type Packages struct {
			XMLName  xml.Name  `xml:"packages"`
			Packages []Package `xml:"package"`
		}
		type App struct {
			XMLName     xml.Name `xml:"app"`
			AppID       string   `xml:"appid,attr"`
			UpdateCheck UpdateCheck
			Version     string   `xml:"version,attr"`
			Packages    Packages `xml:"packages"`
		}
		type RequestWrapper struct {
			XMLName  xml.Name `xml:"request"`
			App      []App    `xml:"app"`
			Protocol string   `xml:"protocol,attr"`
		}

		request := RequestWrapper{}
		err := d.DecodeElement(&request, &start)
		if err != nil {
			return err
		}

		for _, app := range request.App {
			fp := ""
			if len(app.Packages.Packages) > 0 {
				fp = app.Packages.Packages[0].FP
			}
			apps = append(apps, extension.Extension{
				ID:      app.AppID,
				FP:      fp,
				Version: app.Version,
			})
		}

		if request.Protocol != protocolVersion {
			return fmt.Errorf("request version: %v not supported by %v handler", request.Protocol, protocolVersion)
		}
	} else if protocolVersion == "3.1" {
		type App struct {
			XMLName     xml.Name `xml:"app"`
			AppID       string   `xml:"appid,attr"`
			FP          string   `xml:"fp,attr"`
			UpdateCheck UpdateCheck
			Version     string `xml:"version,attr"`
		}
		type RequestWrapper struct {
			XMLName  xml.Name `xml:"request"`
			App      []App    `xml:"app"`
			Protocol string   `xml:"protocol,attr"`
		}

		request := RequestWrapper{}
		err := d.DecodeElement(&request, &start)
		if err != nil {
			return err
		}

		for _, app := range request.App {
			apps = append(apps, extension.Extension{
				ID:      app.AppID,
				FP:      app.FP,
				Version: app.Version,
			})
		}

		if request.Protocol != protocolVersion {
			return fmt.Errorf("request version: %v not supported by %v handler", request.Protocol, protocolVersion)
		}
	} else {
		return fmt.Errorf("unsupported protocol version: %s", protocolVersion)
	}

	*r = apps
	return nil
}

// FilterForUpdates filters the request down to only the extensions that are being checked,
// and only the ones that we have updates for.
func (r *Request) FilterForUpdates(allExtensionsMap *extension.ExtensionsMap) extension.UpdateResponse {
	updateRequest := extension.UpdateRequest(*r)
	return updateRequest.FilterForUpdates(allExtensionsMap)
}

// AsExtensionRequest converts v3.Request to extension.UpdateRequest
func (r *Request) AsExtensionRequest() extension.UpdateRequest {
	return extension.UpdateRequest(*r)
}
