package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var apiUrlTempl string = "https://%v-aiplatform.googleapis.com/v1/projects/%v/locations/%v/publishers/google/models/%v:predict"

type Instance struct {
	Prompt string `json:"prompt"`
}

type Parameters struct {
	SampleCount int64 `json:"sampleCount"`
}

type RequestValues struct {
	Instances  []Instance `json:"instances"`
	Parameters Parameters `json:"parameters"`
}

type Response struct {
	Predictions      []Prediction `json:"predictions"`
	ModelId          string       `json:"deployedModelId"`
	Model            string       `json:"model"`
	ModelDisplayName string       `json:"modelDisplayName"`
	VersionId        string       `json:"modelVersionId"`
}

type Prediction struct {
	Bytes    string `json:"bytesBase64Encoded"`
	MimeType string `json:"mimeType"`
}

var (
	PROJECT_LOCATION string = ""
	PROJECT_ID       string = ""
	TOKEN            string = ""
)

func main() {
	url := makeApiUrl(PROJECT_LOCATION, PROJECT_ID, "imagegeneration@005")
	client := http.DefaultClient
	r := formatRequest(url, "4k, bokeh, front lit, portrait, dark elf wizard eating a churro", TOKEN)
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	response := decodeBody(resp.Body)
	writeRespToDisk(response)
}

func decodeBody(body io.ReadCloser) Response {
	s, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	resp := Response{}
	err = json.Unmarshal(s, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func writeRespToDisk(resp Response) {
	bespokeDir, err := os.MkdirTemp("", "bespoke")
	if err != nil {
		panic(err)
	}

	for i, v := range resp.Predictions {
		bespokeFile, err := os.Create(filepath.Join(bespokeDir, fmt.Sprintf("bespokeImg%v.png", i)))
		if err != nil {
			panic(err)
		}
		defer bespokeFile.Close()
		d, err := base64.StdEncoding.DecodeString(v.Bytes)
		if err != nil {
			panic(err)
		}
		bespokeFile.Write(d)
		fmt.Printf("File Written : %v\n", bespokeFile.Name())
	}
}

func formatRequest(url string, prompt string, token string) *http.Request {
	i := Instance{Prompt: prompt}
	rv := RequestValues{Parameters: Parameters{SampleCount: 4}, Instances: []Instance{i}}
	b, err := json.Marshal(rv)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(b)

	r, err := http.NewRequest("POST", url, buf)
	if err != nil {
		panic(err)
	}
	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	return r
}

func makeApiUrl(location string, project_id string, model_version string) string {
	return fmt.Sprintf(apiUrlTempl, location, project_id, location, model_version)
}
