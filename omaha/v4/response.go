package v4

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/brave/go-update/extension"
	"github.com/go-playground/validator/v10"
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
		URL string `json:"url" validate:"required"`
	}
	type Out struct {
		SHA256 string `json:"sha256" validate:"required"`
		Size   uint64 `json:"size,omitempty" validate:"gt=0"`
	}
	type In struct {
		SHA256 string `json:"sha256" validate:"required"`
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

	// Create validator instance
	validate := validator.New()

	for _, ext := range *r {
		// Check if SHA256 is empty
		if ext.SHA256 == "" {
			return nil, fmt.Errorf("extension %s has empty SHA256", ext.ID)
		}

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
					// Check if hashdiff is empty
					if patchInfo.Hashdiff == "" {
						return nil, fmt.Errorf("extension %s has empty Hashdiff", ext.ID)
					}

					fpPrefix := ext.FP
					if len(ext.FP) >= 8 {
						fpPrefix = ext.FP[:8]
					}
					diffPipelineID := "puff_diff_" + fpPrefix
					patchURL := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" +
						ext.ID + "/patches/" + ext.SHA256 + "/" + ext.FP + ".puff"

					// Create the Out struct for diff pipeline
					diffOut := &Out{
						SHA256: patchInfo.Hashdiff,
						Size:   normalizeSize(uint64(patchInfo.Sizediff)), // Use Sizediff if available, normalize for validation
					}
					// Validate the Out struct
					if err := validate.Struct(diffOut); err != nil {
						return nil, fmt.Errorf("diff validation failed for extension %s: %v", ext.ID, err)
					}

					// Create and validate URLs for diff pipeline
					diffURLs := []URL{{URL: patchURL}}
					for i, u := range diffURLs {
						if err := validate.Struct(u); err != nil {
							return nil, fmt.Errorf("diff URL validation failed for extension %s (URL %d): %v", ext.ID, i, err)
						}
					}

					// Create and validate the In struct for the previous field
					previousIn := &In{SHA256: ext.FP}
					if err := validate.Struct(previousIn); err != nil {
						return nil, fmt.Errorf("previous validation failed for extension %s: %v", ext.ID, err)
					}

					// Create and validate the In struct for crx3
					crx3In := &In{SHA256: ext.SHA256}
					if err := validate.Struct(crx3In); err != nil {
						return nil, fmt.Errorf("crx3 In validation failed for extension %s: %v", ext.ID, err)
					}

					diffPipeline := Pipeline{
						PipelineID: diffPipelineID,
						Operations: []Operation{
							{
								Type: "download",
								Out:  diffOut,
								URLs: diffURLs,
							},
							{
								Type:     "puff",
								Previous: previousIn,
							},
							{
								Type: "crx3",
								In:   crx3In,
							},
						},
					}

					app.UpdateCheck.Pipelines = append(app.UpdateCheck.Pipelines, diffPipeline)
				}
			}

			// Add full pipeline as fallback (always add as the last pipeline)
			// Create Out struct with normalized size
			out := &Out{
				SHA256: ext.SHA256,
				Size:   normalizeSize(ext.Size),
			}
			// Validate the Out struct
			if err := validate.Struct(out); err != nil {
				return nil, fmt.Errorf("validation failed for extension %s: %v", ext.ID, err)
			}

			// Create URLs array
			urls := []URL{{URL: url}}
			// Validate individual URL struct
			for i, u := range urls {
				if err := validate.Struct(u); err != nil {
					return nil, fmt.Errorf("URL validation failed for extension %s (URL %d): %v", ext.ID, i, err)
				}
			}

			// Create and validate the In struct for crx3 in the main pipeline
			mainCrx3In := &In{SHA256: ext.SHA256}
			if err := validate.Struct(mainCrx3In); err != nil {
				return nil, fmt.Errorf("main crx3 In validation failed for extension %s: %v", ext.ID, err)
			}

			pipeline := Pipeline{
				PipelineID: "direct_full",
				Operations: []Operation{
					{
						Type: "download",
						Out:  out,
						URLs: urls,
					},
					{
						Type: "crx3",
						In:   mainCrx3In,
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

func normalizeSize(size uint64) uint64 {
	if size == 0 {
		return 1
	}
	return size
}
