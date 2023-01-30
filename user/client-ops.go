package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bofrim/gorch/hook"
)

// Function for sending a get request to an orchestrator
func GetNodes(addr string, port int) ([]byte, error) {
	// Prepare the request
	url := fmt.Sprintf("http://%s:%d/nodes", addr, port)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Do the request
	client := &http.Client{}
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

func RunAction(addr string, port int, node string, action string, data map[string]string) error {
	url := fmt.Sprintf("http://%s:%d/%s/action/%s", addr, port, node, action)
	return DoPostRequest(url, data)
}

func StreamAction(addr string, port int, node string, streamPort int, action string, data map[string]string) error {
	url := fmt.Sprintf("http://%s:%d/%s/action/%s", addr, port, node, action)
	data["stream_addr"] = "loopback"
	data["stream_port"] = fmt.Sprintf("%d", streamPort)
	postErr := DoPostRequest(url, data)
	if postErr != nil {
		return postErr
	}

	// Start a hook listener
	h := hook.HookListener{}
	return h.Listen(streamPort)

}

func RequestData(addr string, port int, node string, path string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/%s/data/%s", addr, port, node, path)
	return DoGetRequest(url)
}

func RequestDataList(addr string, port int, node string, path string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/%s/list/%s", addr, port, node, path)
	fmt.Println("Requesting data list from: " + url)
	return DoGetRequest(url)
}

func DoGetRequest(url string) ([]byte, error) {
	// Prepare the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\n", resp.Status)
		return nil, fmt.Errorf("request not OK: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func DoPostRequest(url string, data map[string]string) error {
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
	req.Close = true

	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Bad request: %s\n", resp.Status)
		return fmt.Errorf("request not OK: %d", resp.StatusCode)
	}

	// Process the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n\n", body)
	return nil
}
