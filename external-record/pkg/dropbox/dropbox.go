package dropbox

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	dropboxTokenURL     = "https://api.dropboxapi.com/oauth2/token"
	startSessionURL     = "https://content.dropboxapi.com/2/files/upload_session/start"
	appendURL           = "https://content.dropboxapi.com/2/files/upload_session/append_v2"
	finishURL           = "https://content.dropboxapi.com/2/files/upload_session/finish"
	createSharedLinkURL = "https://api.dropboxapi.com/2/sharing/create_shared_link"
)

type Dropbox interface {
	GetAccessToken() (string, error)
	UploadFile(token, filePath, newFileName string) (string, error)
}

type dropboxClient struct {
	AppKey       string
	AppSecret    string
	RefreshToken string
}

func (d *dropboxClient) GetAccessToken() (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", d.AppKey, d.AppSecret)))

	data := url.Values{}
	data.Set("refresh_token", d.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest(http.MethodPost, dropboxTokenURL, ioutil.NopCloser(strings.NewReader(data.Encode())))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("error: %s", body)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	_, err = json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	if accessToken, ok := result["access_token"].(string); ok {
		return accessToken, nil
	}

	return "", fmt.Errorf("access token not found in response")
}

func (d *dropboxClient) UploadFile(dbxAccessToken, filePath, newFileName string) (string, error) {
	fileStats, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting file stats: %v", err)
	}
	fileSize := fileStats.Size()
	// Step 1: Start an upload session
	startSessionReq, err := http.NewRequest("POST", startSessionURL, bytes.NewReader([]byte{}))
	if err != nil {
		return "", fmt.Errorf("error creating start session request: %v", err)
	}
	startSessionReq.Header.Set("Authorization", "Bearer "+dbxAccessToken)
	startSessionReq.Header.Set("Content-Type", "application/octet-stream")

	startSessionResp, err := http.DefaultClient.Do(startSessionReq)
	if err != nil {
		return "", fmt.Errorf("error starting upload session: %v", err)
	}
	defer startSessionResp.Body.Close()

	if startSessionResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error starting upload session with status: %d", startSessionResp.StatusCode)
	}

	var startSessionRespData struct {
		SessionID string `json:"session_id"`
	}
	err = json.NewDecoder(startSessionResp.Body).Decode(&startSessionRespData)
	if err != nil {
		return "", fmt.Errorf("error decoding start session response: %v", err)
	}

	sessionID := startSessionRespData.SessionID
	cursor := 0
	chunkSize := 10 * 1024 * 1024 // 10 MB chunk size

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	for {
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading file chunk: %v", err)
		}

		// Step 2: Append chunk to upload session
		req, err := http.NewRequest("POST", appendURL, bytes.NewReader(chunk[:n]))
		if err != nil {
			return "", fmt.Errorf("error creating append request: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+dbxAccessToken)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Dropbox-API-Arg", fmt.Sprintf(`{"cursor":{"session_id":"%s","offset":%d}}`, sessionID, cursor))

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("error appending chunk: %v", err)
		}
		cursor += n

	}

	parts := strings.Split(filePath, "/")[len(strings.Split(filePath, "/"))-2:] // Get the file extension
	folder := parts[0]
	fileName := parts[1]
	extension := filepath.Ext(fileName)
	newFilePath := folder + "/" + newFileName + extension

	// Step 3: Finish upload session
	finishArg := fmt.Sprintf(`{"cursor":{"session_id":"%s","offset":%d},"commit":{"path":"/bbb-%s/%s","mode":"add","autorename":true,"mute":false}}`, sessionID, fileSize, time.Now().Format("2006-01-02"), newFilePath)
	finishReq, err := http.NewRequest("POST", finishURL, bytes.NewReader([]byte{}))
	if err != nil {
		return "", fmt.Errorf("error creating finish request: %v", err)
	}
	finishReq.Header.Set("Authorization", "Bearer "+dbxAccessToken)
	finishReq.Header.Set("Content-Type", "application/octet-stream")
	finishReq.Header.Set("Dropbox-API-Arg", finishArg)

	_, err = http.DefaultClient.Do(finishReq)
	if err != nil {
		return "", fmt.Errorf("error finishing upload session: %v", err)
	}
	sharedLinkReqBody := fmt.Sprintf(`{"path": "/bbb-%s/%s"}`, time.Now().Format("2006-01-02"), newFilePath)
	sharedLinkReq, err := http.NewRequest("POST", createSharedLinkURL, bytes.NewBuffer([]byte(sharedLinkReqBody)))
	if err != nil {
		return "", fmt.Errorf("error creating shared link request: %v", err)
	}
	sharedLinkReq.Header.Set("Authorization", "Bearer "+dbxAccessToken)
	sharedLinkReq.Header.Set("Content-Type", "application/json")

	sharedLinkResp, err := http.DefaultClient.Do(sharedLinkReq)
	if err != nil {
		return "", fmt.Errorf("error getting shared link: %v", err)
	}
	defer sharedLinkResp.Body.Close()

	if sharedLinkResp.StatusCode != http.StatusOK {
		// Read and log the response body for debugging
		bodyBytes, _ := io.ReadAll(sharedLinkResp.Body)
		bodyString := string(bodyBytes)
		return "", fmt.Errorf("error creating shared link with status: %d, response: %s", sharedLinkResp.StatusCode, bodyString)
	}

	var sharedLinkRespData struct {
		URL string `json:"url"`
	}
	err = json.NewDecoder(sharedLinkResp.Body).Decode(&sharedLinkRespData)
	if err != nil {
		return "", fmt.Errorf("error decoding shared link response: %v", err)
	}

	// Modify the link to enable direct download
	publicURL := strings.Replace(sharedLinkRespData.URL, "dl=0", "dl=1", 1)

	return publicURL, nil
}

func NewDropBox(appKey, appSecret, refreshToken string) Dropbox {
	return &dropboxClient{
		AppKey:       appKey,
		AppSecret:    appSecret,
		RefreshToken: refreshToken,
	}
}
