package v4

import (
	"encoding/json"
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
		Size   uint64 `json:"size,omitempty" validate:"gt=0"`
	}
	type In struct {
		SHA256 string `json:"sha256"`
	}
	type Operation struct {
		Type     string `json:"type"`
		Out      *Out   `json:"out,omitempty"`
		In       *In    `json:"in,omitempty"`
		URLs     []URL  `json:"urls,omitempty"`
		Previous *In    `json:"previous,omitempty"`
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

			// Initialize pipelines array
			app.UpdateCheck.Pipelines = []Pipeline{}

			// Add diff pipeline if patch is available (diff pipeline should come first)
			if ext.FP != "" && ext.PatchList != nil {
				if patchInfo, ok := ext.PatchList[ext.FP]; ok {
					fpPrefix := ext.FP
					if len(ext.FP) >= 8 {
						fpPrefix = ext.FP[:8]
					}
					diffPipelineID := "puff_diff_" + fpPrefix
					patchURL := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" +
						ext.ID + "/patches/" + ext.SHA256 + "/" + ext.FP + ".puff"

					diffPipeline := Pipeline{
						PipelineID: diffPipelineID,
						Operations: []Operation{
							{
								Type: "download",
								Out:  &Out{SHA256: patchInfo.Hashdiff},
								URLs: []URL{{URL: patchURL}},
							},
							{
								Type:     "puff",
								Previous: &In{SHA256: ext.FP},
							},
							{
								Type: "crx3",
								In:   &In{SHA256: ext.SHA256},
							},
						},
					}

					app.UpdateCheck.Pipelines = append(app.UpdateCheck.Pipelines, diffPipeline)
				}
			}

			// Add full pipeline as fallback (always add as the last pipeline)
			pipeline := Pipeline{
				PipelineID: "direct_full",
				Operations: []Operation{
					{
						Type: "download",
						Out:  &Out{SHA256: ext.SHA256, Size: normalizeSize(ext.Size)},
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

// normalizeSize ensures Size is greater than 0
func normalizeSize(size uint64) uint64 {
	if size == 0 {
		return 1
	}
	return size
}
