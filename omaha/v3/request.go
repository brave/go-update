package v3

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/brave/go-update/extension"
	"github.com/go-playground/validator/v10"
)

// UpdateRequest represents an Omaha v3 update request
type UpdateRequest []extension.Extension

// UnmarshalJSON decodes the update server request JSON data
func (r *UpdateRequest) UnmarshalJSON(b []byte) error {
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
		Protocol string `json:"protocol" validate:"required"`
	}
	type JSONRequest struct {
		Request RequestWrapper `json:"request" validate:"required"`
	}

	request := JSONRequest{}
	if err := json.Unmarshal(b, &request); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(request); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		return fmt.Errorf("request validation failed: %v", validationErrors)
	}

	*r = UpdateRequest{}
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

	return nil
}

// UnmarshalXML decodes the update server request XML data
func (r *UpdateRequest) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	// Common XML elements
	type UpdateCheck struct {
		XMLName xml.Name `xml:"updatecheck"`
	}

	// Check request for protocol version
	var protocol string
	for _, attr := range start.Attr {
		if attr.Name.Local == "protocol" {
			protocol = attr.Value
			break
		}
	}

	// Version-specific types
	var apps []extension.Extension

	if protocol == "3.0" {
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
	} else if protocol == "3.1" {
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
	} else {
		// Default to the simplest structure
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
	}

	*r = apps
	return nil
}
