package v4

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"

	"github.com/brave/go-update/extension"
)

// GetElapsedDays calculates elapsed days since Jan 1, 2007
var GetElapsedDays = func() int {
	startDate := time.Date(2007, 1, 1, 0, 0, 0, 0, time.UTC)
	return int(time.Now().UTC().Sub(startDate).Hours() / 24)
}

// UpdateResponse represents an Omaha v4 update response
type UpdateResponse []extension.Extension

// GetUpdateStatus determines the update status based on extension data
func GetUpdateStatus(extension extension.Extension) string {
	if extension.Status == "noupdate" {
		return "noupdate"
	}
	return "ok"
}

// MarshalJSON encodes the extension list into response JSON
func (r *UpdateResponse) MarshalJSON() ([]byte, error) {
	type URL struct {
		URL string `json:"url"`
	}
	type Out struct {
		SHA256 string `json:"sha256"`
	}
	type In struct {
		SHA256 string `json:"sha256"`
	}
	type Operation struct {
		Type string `json:"type"`
		Out  *Out   `json:"out,omitempty"`
		In   *In    `json:"in,omitempty"`
		URLs []URL  `json:"urls,omitempty"`
	}
	type Pipeline struct {
		PipelineID string      `json:"pipeline_id"`
		Operations []Operation `json:"operations"`
	}
	type UpdateCheck struct {
		Status      string     `json:"status"`
		NextVersion string     `json:"nextversion,omitempty"`
		Pipelines   []Pipeline `json:"pipelines,omitempty"`
	}
	type DayStart struct {
		ElapsedDays int `json:"elapsed_days"`
	}
	type App struct {
		AppID       string      `json:"appid"`
		Status      string      `json:"status"`
		UpdateCheck UpdateCheck `json:"updatecheck"`
	}
	type ResponseWrapper struct {
		Protocol string   `json:"protocol"`
		DayStart DayStart `json:"daystart"`
		Apps     []App    `json:"apps"`
	}
	type JSONResponse struct {
		Response ResponseWrapper `json:"response"`
	}

	// Calculate elapsed days since Jan 1, 2007
	elapsedDays := GetElapsedDays()

	response := ResponseWrapper{
		Protocol: "4.0",
		DayStart: DayStart{
			ElapsedDays: elapsedDays,
		},
	}

	for _, ext := range *r {
		app := App{AppID: ext.ID, Status: "ok"}
		updateStatus := GetUpdateStatus(ext)
		app.UpdateCheck = UpdateCheck{Status: updateStatus}

		if updateStatus == "ok" {
			app.UpdateCheck.NextVersion = ext.Version

			// Create pipeline with operations
			extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
			url := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName

			pipeline := Pipeline{
				PipelineID: "direct_full",
				Operations: []Operation{
					{
						Type: "download",
						Out:  &Out{SHA256: ext.SHA256},
						URLs: []URL{{URL: url}},
					},
					{
						Type: "crx3",
						In:   &In{SHA256: ext.SHA256},
					},
				},
			}

			app.UpdateCheck.Pipelines = append(app.UpdateCheck.Pipelines, pipeline)
		}

		response.Apps = append(response.Apps, app)
	}

	jsonResponse := JSONResponse{
		Response: response,
	}

	return json.Marshal(jsonResponse)
}

// WebStoreResponse represents a web store update response
type WebStoreResponse []extension.Extension

// MarshalJSON encodes the extension list into web store response JSON
func (r *WebStoreResponse) MarshalJSON() ([]byte, error) {
	type UpdateCheck struct {
		Status   string `json:"status"`
		Codebase string `json:"codebase"`
		Version  string `json:"version"`
		SHA256   string `json:"hash_sha256"`
	}
	type App struct {
		AppID       string      `json:"appid"`
		Status      string      `json:"status"`
		UpdateCheck UpdateCheck `json:"updatecheck"`
	}
	type GUpdate struct {
		Protocol string `json:"protocol"`
		Server   string `json:"server"`
		Apps     []App  `json:"app"`
	}
	type JSONGUpdate struct {
		GUpdate GUpdate `json:"gupdate"`
	}
	response := GUpdate{}
	response.Protocol = "4.0"
	response.Server = "prod"

	for _, ext := range *r {
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  ext.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   ext.SHA256,
				Version:  ext.Version,
				Codebase: "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	jsonGupdate := JSONGUpdate{}
	jsonGupdate.GUpdate = response

	return json.Marshal(jsonGupdate)
}

// MarshalXML encodes the extension list into web store response XML
func (r *WebStoreResponse) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	type UpdateCheck struct {
		XMLName  xml.Name `xml:"updatecheck"`
		Status   string   `xml:"status,attr"`
		Codebase string   `xml:"codebase,attr"`
		Version  string   `xml:"version,attr"`
		SHA256   string   `xml:"hash_sha256,attr"`
	}
	type App struct {
		XMLName     xml.Name `xml:"app"`
		AppID       string   `xml:"appid,attr"`
		Status      string   `xml:"status,attr"`
		UpdateCheck UpdateCheck
	}
	type GUpdate struct {
		XMLName  xml.Name `xml:"gupdate"`
		Protocol string   `xml:"protocol,attr"`
		Server   string   `xml:"server,attr"`
		Apps     []App
	}
	response := GUpdate{}
	response.Protocol = "4.0"
	response.Server = "prod"

	for _, ext := range *r {
		extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
		app := App{
			AppID:  ext.ID,
			Status: "ok",
			UpdateCheck: UpdateCheck{
				Status:   "ok",
				SHA256:   ext.SHA256,
				Version:  ext.Version,
				Codebase: "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName,
			},
		}
		response.Apps = append(response.Apps, app)
	}
	e.Indent("", "    ")
	err := e.EncodeElement(response, xml.StartElement{Name: xml.Name{Local: "gupdate"}})
	return err
}
