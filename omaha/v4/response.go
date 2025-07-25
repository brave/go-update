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
	// Return the existing status if already set (indicates no update available or an error)
	if extension.Status != "" {
		return extension.Status
	}
	// Unassigned status implies an available update
	return "ok"
}

// MarshalJSON encodes the extension list into response JSON
func (r *UpdateResponse) MarshalJSON() ([]byte, error) {
	type URL struct {
		URL string `json:"url" validate:"required"`
	}
	type Out struct {
		SHA256 string `json:"sha256" validate:"required"`
	}
	type In struct {
		SHA256 string `json:"sha256" validate:"required"`
	}
	type Operation struct {
		Type     string `json:"type" validate:"required,oneof=download puff crx3"`
		Out      *Out   `json:"out,omitempty" validate:"omitempty,required_if=Type download,required_if=Type puff"`
		In       *In    `json:"in,omitempty" validate:"omitempty,required_if=Type crx3"`
		URLs     []URL  `json:"urls,omitempty" validate:"omitempty,required_if=Type download,dive"`
		Previous *In    `json:"previous,omitempty" validate:"omitempty,required_if=Type puff"`
		Size     uint64 `json:"size,omitempty" validate:"omitempty,required_if=Type download,gt=0"`
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
		updateStatus := GetUpdateStatus(ext)
		app := App{
			AppID:       ext.ID,
			Status:      "ok",
			UpdateCheck: UpdateCheck{Status: updateStatus},
		}

		// Further processing makes sense only if there is an update available
		if updateStatus == "ok" {
			// Check if SHA256 is empty
			if ext.SHA256 == "" {
				return nil, fmt.Errorf("extension %s has empty SHA256", ext.ID)
			}

			app.UpdateCheck.NextVersion = ext.Version

			// Create pipeline with operations
			extensionName := "extension_" + strings.Replace(ext.Version, ".", "_", -1) + ".crx"
			url := "https://" + extension.GetS3ExtensionBucketHost(ext.ID) + "/release/" + ext.ID + "/" + extensionName

			// Initialize pipelines array
			app.UpdateCheck.Pipelines = []Pipeline{}

			mainCrx3Out := &Out{
				SHA256: ext.SHA256,
			}
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
					}

					// Create URLs for diff pipeline
					diffURLs := []URL{{URL: patchURL}}

					// Create In structs
					previousIn := &In{SHA256: ext.FP}
					crx3In := &In{SHA256: ext.SHA256}

					// Create operations for diff pipeline
					diffDownloadOp := Operation{
						Type: "download",
						Out:  diffOut,
						URLs: diffURLs,
						Size: normalizeSize(uint64(patchInfo.Sizediff)),
					}

					puffOp := Operation{
						Type:     "puff",
						Previous: previousIn,
						Out:      mainCrx3Out,
					}

					crx3Op := Operation{
						Type: "crx3",
						In:   crx3In,
					}

					// Validate all operations
					for _, op := range []Operation{diffDownloadOp, puffOp, crx3Op} {
						if err := validate.Struct(op); err != nil {
							return nil, fmt.Errorf("%s operation validation failed for extension %s: %v", op.Type, ext.ID, err)
						}
					}

					diffPipeline := Pipeline{
						PipelineID: diffPipelineID,
						Operations: []Operation{
							diffDownloadOp,
							puffOp,
							crx3Op,
						},
					}

					app.UpdateCheck.Pipelines = append(app.UpdateCheck.Pipelines, diffPipeline)
				}
			}

			// Add full pipeline as fallback (always add as the last pipeline)
			urls := []URL{{URL: url}}
			mainCrx3In := &In{SHA256: ext.SHA256}

			// Create operations for main pipeline
			mainDownloadOp := Operation{
				Type: "download",
				Out:  mainCrx3Out,
				URLs: urls,
				Size: normalizeSize(ext.Size),
			}

			mainCrx3Op := Operation{
				Type: "crx3",
				In:   mainCrx3In,
			}

			// Validate all operations in the main pipeline
			for _, op := range []Operation{mainDownloadOp, mainCrx3Op} {
				if err := validate.Struct(op); err != nil {
					return nil, fmt.Errorf("%s operation validation failed for extension %s: %v", op.Type, ext.ID, err)
				}
			}

			pipeline := Pipeline{
				PipelineID: "direct_full",
				Operations: []Operation{
					mainDownloadOp,
					mainCrx3Op,
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
