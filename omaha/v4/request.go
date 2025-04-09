package v4

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

	"github.com/brave/go-update/extension"
	"github.com/go-playground/validator/v10"
)

// UpdateRequest represents an Omaha v4 update request
type UpdateRequest []extension.Extension

// UnmarshalJSON decodes the update server request JSON data
func (r *UpdateRequest) UnmarshalJSON(b []byte) error {
	type CachedItem struct {
		SHA256 string `json:"sha256"`
	}
	type App struct {
		AppID       string       `json:"appid"`
		Version     string       `json:"version"`
		CachedItems []CachedItem `json:"cached_items"`
	}
	type RequestWrapper struct {
		OS           string `json:"@os"`
		Updater      string `json:"@updater"`
		Apps         []App  `json:"apps"`
		Protocol     string `json:"protocol" validate:"required"`
		AcceptFormat string `json:"acceptformat"`
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
	for _, app := range request.Request.Apps {
		fp := ""
		if len(app.CachedItems) > 0 {
			fp = app.CachedItems[0].SHA256
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

	// Only process protocol v4.0
	if protocol != "4.0" {
		return fmt.Errorf("unsupported protocol version: %s", protocol)
	}

	type CachedItem struct {
		XMLName xml.Name `xml:"cacheditem"`
		SHA256  string   `xml:"sha256,attr"`
	}
	type CachedItems struct {
		XMLName     xml.Name     `xml:"cacheditems"`
		CachedItems []CachedItem `xml:"cacheditem"`
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		Version     string   `xml:"version,attr"`
		UpdateCheck UpdateCheck
		CachedItems CachedItems `xml:"cacheditems"`
	}
	type RequestWrapper struct {
		XMLName      xml.Name `xml:"request"`
		Apps         []App    `xml:"app"`
		Protocol     string   `xml:"protocol,attr"`
		AcceptFormat string   `xml:"acceptformat,attr"`
	}

	request := RequestWrapper{}
	err := d.DecodeElement(&request, &start)
	if err != nil {
		return err
	}

	*r = UpdateRequest{}
	for _, app := range request.Apps {
		fp := ""
		if len(app.CachedItems.CachedItems) > 0 {
			fp = app.CachedItems.CachedItems[0].SHA256
		}
		*r = append(*r, extension.Extension{
			ID:      app.AppID,
			FP:      fp,
			Version: app.Version,
		})
	}

	return nil
}
