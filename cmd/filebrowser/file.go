/*
* Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
*
* SPDX-License-Identifier: Apache-2.0
*
*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at
*
*       http://www.apache.org/licenses/LICENSE-2.0
*
*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
*
 */

package filebrowser

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-openapi/runtime"
	"softwareag.com/client"
	"softwareag.com/client/browser"
)

// Locations list all available file location
func Locations(clientInstance *client.AdabasAdmin, auth runtime.ClientAuthInfoWriter) error {
	params := browser.NewBrowseListParams()
	resp, err := clientInstance.Browser.BrowseList(params, auth)
	if err != nil {
		switch err.(type) {
		case *browser.BrowseListBadRequest:
			response := err.(*browser.BrowseListBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf(" Name                              | Location\n")
	fmt.Printf("-----------------------------------|----------------------------------------\n")
	for _, d := range resp.Payload.Directories {
		fmt.Printf(" %-33s | %s\n", d.Name, d.Location)
	}
	return nil
}

// List list the files of an specific file location
func List(clientInstance *client.AdabasAdmin, param string, auth runtime.ClientAuthInfoWriter) error {
	params := browser.NewBrowseParams()
	if param == "" {
		fmt.Printf("Parameter option missing\n")
		return fmt.Errorf("Parameter option missing")
	}
	p := strings.Split(param, ":")
	if len(p) != 2 {
		fmt.Printf("Invalid parameter option: %s\n", param)
		fmt.Printf("Need to be of format: <location>:<reference>\n")
		return fmt.Errorf("Need to be of format: <location>:<reference>")
	}
	params.Location = p[0]
	params.File = p[1]
	resp, err := clientInstance.Browser.Browse(params, auth)
	if err != nil {
		switch err.(type) {
		case *browser.BrowseBadRequest:
			response := err.(*browser.BrowseBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Println("Reference : ", resp.Payload.Reference)
	fmt.Println("Location : ", resp.Payload.Location)
	for _, f := range resp.Payload.Content {
		fmt.Printf(" %-20s %-8d %-10s %-10s %-10s\n", f.Name, f.Size, f.Type, f.Modified, f.Created)
	}
	fmt.Println()
	return nil
}

// Download download a file
func Download(clientInstance *client.AdabasAdmin, param string, input string, auth runtime.ClientAuthInfoWriter) error {
	params := browser.NewDownloadFileParams()
	p := strings.Split(param, ":")
	params.Location = p[0]
	params.File = p[1]
	file, err := os.OpenFile(input, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = clientInstance.Browser.DownloadFile(params, auth, file)
	if err != nil {
		switch err.(type) {
		case *browser.DownloadFileBadRequest:
			response := err.(*browser.DownloadFileBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	return nil
}

// Upload upload file
func Upload(clientInstance *client.AdabasAdmin, param string, input string, auth runtime.ClientAuthInfoWriter) error {
	params := browser.NewUploadFileParams()
	p := strings.Split(param, ":")
	params.Location = p[0]
	params.File = p[1]
	file, ferr := os.OpenFile(input, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if ferr != nil {
		return ferr
	}
	params.UploadFile = runtime.NamedReader(input, file)
	defer file.Close()
	resp, err := clientInstance.Browser.UploadFile(params, auth)
	if err != nil {
		switch err.(type) {
		case *browser.UploadFileBadRequest:
			response := err.(*browser.UploadFileBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println("Upload ", resp.Payload.Status.Message)
	return nil
}
