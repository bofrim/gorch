package user

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bofrim/gorch/hook"
)

// Function for sending a get request to an orchestrator
func GetNodes(addr string) ([]byte, error) {
	// Prepare the request
	url := fmt.Sprintf("https://%s/nodes", addr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Do the request
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad request: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func RunAction(addr string, node string, action string, data map[string]interface{}, headers map[string]string) error {
	url := fmt.Sprintf("https://%s/%s/action/%s", addr, node, action)
	return DoPostRequest(url, data, headers)
}

func StreamAction(addr string, node string, streamPort int, action string, data map[string]interface{}, headers map[string]string) error {
	url := fmt.Sprintf("https://%s/%s/action/%s", addr, node, action)
	data["stream_addr"] = "loopback"
	data["stream_port"] = fmt.Sprintf("%d", streamPort)
	postErr := DoPostRequest(url, data, headers)
	if postErr != nil {
		return postErr
	}

	// Start a hook listener
	h := hook.NewHookListener()
	return h.Listen(streamPort)
}

func RequestData(addr string, node string, path string, headers map[string]string) ([]byte, error) {
	url := fmt.Sprintf("https://%s/%s/data/%s", addr, node, path)
	return DoGetRequest(url, headers)
}

func RequestDataList(addr string, node string, path string, headers map[string]string) ([]byte, error) {
	url := fmt.Sprintf("https://%s/%s/list/%s", addr, node, path)
	fmt.Println("Requesting data list from: " + url)
	return DoGetRequest(url, headers)
}

func DoGetRequest(url string, headers map[string]string) ([]byte, error) {
	// Prepare the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Do the request
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\n", resp.Status)
		return nil, fmt.Errorf("get request not OK: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func DoPostRequest(url string, data map[string]interface{}, headers map[string]string) error {
	serial, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(serial))
	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		fmt.Printf("Setting header: %s: %s\n", k, v)
		req.Header.Set(k, v)
	}
	req.Close = true

	// Do the request
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\nReq:\n%+v\n", resp.Status, req)
		return fmt.Errorf("post request not OK: %d", resp.StatusCode)
	}

	// Process the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n\n", body)
	return nil
}
