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

package database

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-openapi/runtime"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"softwareag.com/client"
	"softwareag.com/client/online"
	"softwareag.com/client/online_offline"
	"softwareag.com/models"
)

// InputList input list
type InputList []string

func (il *InputList) String() string {
	var buffer bytes.Buffer
	for _, x := range *il {
		if buffer.Len() > 0 {
			buffer.WriteString(",")
		}
		buffer.WriteString(x)
	}
	return buffer.String()
}

// Set set the corresponding list entry
func (il *InputList) Set(v string) error {
	*il = append(*il, v)
	return nil
}

// Files list database files
func Files(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseFilesParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.OnlineOffline.GetDatabaseFiles(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseFilesBadRequest:
			response := err.(*online_offline.GetDatabaseFilesBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Printf("Database %03d files:\n", dbid)
	p.Println()
	p.Println("File  Name                Record count ")
	p.Println("----  ------------------- ------------ ")
	for _, f := range resp.Payload.Files {
		s := ""
		if f.IsLob > 0 {
			s = fmt.Sprintf("Lobfile of %d", f.IsLob)
		}
		if f.IsLobRoot > 0 {
			s = fmt.Sprintf("Lob is %d", f.IsLobRoot)
		}
		p.Printf(" %03d  %-20s %10d %s\n", f.FileNr, f.Name, f.RecordCount, s)
	}
	return nil
}

// File get file information
func File(clientInstance *client.AdabasAdmin, dbid int, fnr int, para string, auth runtime.ClientAuthInfoWriter) error {
	if dbid < 1 {
		fmt.Println("Please add option -dbid Adabas database id")
		return fmt.Errorf("Please add option -dbid Adabas database id")
	}
	if fnr < 1 {
		fmt.Println("Please add option -fnr Adabas file number")
		return fmt.Errorf("Please add option -fnr Adabas file number")
	}
	if para != "" {
		params := online.NewPutAdabasFileParameterParams()
		x := strings.Split(para, ",")
		for _, pa := range x {
			p := strings.Split(pa, "=")
			fmt.Println("Name:" + p[0])
			fmt.Println("Value:" + p[1])
			var pgm bool
			var ir bool
			var sr bool
			switch p[0] {
			case "pgmRefresh":
				b, err := strconv.ParseBool(p[1])
				if err != nil {
					fmt.Println("PGM refresh parameter invalid: ", p[1])
					return err
				}
				pgm = b
				params.Pgmrefresh = &pgm
			case "isnReusage":
				b, err := strconv.ParseBool(p[1])
				if err != nil {
					fmt.Println("ISN reusage parameter invalid: ", p[1])
					return err
				}
				ir = b
				params.Isnreusage = &ir
			case "spaceReusage":
				b, err := strconv.ParseBool(p[1])
				if err != nil {
					fmt.Println("Space reusage parameter invalid: ", p[1])
					return err
				}
				sr = b
				params.Spacereusage = &sr
			default:
				fmt.Println("Unknown parameter:", p[0])
				return fmt.Errorf("Unkown Parameter %s", p[0])
			}
		}

		params.Dbid = float64(dbid)
		params.FileOperation = strconv.Itoa(fnr)
		resp, errp := clientInstance.Online.PutAdabasFileParameter(params, auth)
		if errp != nil {
			switch errp.(type) {
			case *online.PutAdabasFileParameterBadRequest:
				response := errp.(*online.PutAdabasFileParameterBadRequest)
				fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
			default:
				fmt.Println("Error:", errp)
			}
			return errp
		}
		fmt.Println("Status: ", resp.Payload.Status.Message)
		return nil
	}
	params := online_offline.NewGetDatabaseFileParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
	params.Dbid = float64(dbid)
	params.FileOperation = strconv.Itoa(fnr)
	resp, err := clientInstance.OnlineOffline.GetDatabaseFile(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseFileBadRequest:
			response := err.(*online_offline.GetDatabaseFileBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Printf("\nDatabase %03d file %03d:\n", dbid, fnr)
	p.Println()
	p.Printf("Name                : %s\n", resp.Payload.File.Name)
	p.Printf("Number              : %d\n", resp.Payload.File.Number)
	p.Printf("Last modification   : %s\n", resp.Payload.File.LastModification)
	p.Printf("Flags               : %s\n", resp.Payload.File.Flags)
	p.Printf("ISN count           : %d\n", resp.Payload.File.IsnCnt)
	p.Printf("Top ISN             : %d\n", resp.Payload.File.TopIsn)
	p.Printf("Maximum ISN         : %d\n", resp.Payload.File.MaxIsn)
	p.Printf("Max.MU Occurence    : %d\n", resp.Payload.File.MaxMuOccurence)
	p.Printf("Padding factor ASSO : %d\n", resp.Payload.File.PaddingFactorAsso)
	p.Printf("Padding factor DATA : %d\n", resp.Payload.File.PaddingFactorData)
	p.Printf("Max.record length   : %d\n", resp.Payload.File.MaxRecordLength)
	p.Printf("Structure level     : %d\n", resp.Payload.File.StructureLevel)
	p.Printf("Root file           : %d\n", resp.Payload.File.RootFile)
	p.Printf("Lob file            : %d\n", resp.Payload.File.LobFile)
	p.Printf("Record count        : %d\n", resp.Payload.File.RecordCount)
	p.Printf("Security info       : %d\n", resp.Payload.File.SecurityInfo)
	p.Printf("AC extents\n")
	for _, e := range resp.Payload.File.ACextents {
		p.Printf(" - First RABN   : %d\n", e.FirstRabn)
		p.Printf("   Last RABN    : %d\n", e.LastRabn)
		p.Printf("   Free or Isn  : %d\n", e.FreeOrIsn)
	}
	p.Printf("DS extents\n")
	for _, e := range resp.Payload.File.DSextents {
		p.Printf(" - First RABN   : %d\n", e.FirstRabn)
		p.Printf("   Last RABN    : %d\n", e.LastRabn)
		p.Printf("   Free or Isn  : %d\n", e.FreeOrIsn)
	}
	p.Printf("NI extents\n")
	for _, e := range resp.Payload.File.NIextents {
		p.Printf(" - First RABN   : %d\n", e.FirstRabn)
		p.Printf("   Last RABN    : %d\n", e.LastRabn)
		p.Printf("   Free or Isn  : %d\n", e.FreeOrIsn)
	}
	p.Printf("UI extents\n")
	for _, e := range resp.Payload.File.UIextents {
		p.Printf(" - First RABN   : %d\n", e.FirstRabn)
		p.Printf("   Last RABN    : %d\n", e.LastRabn)
		p.Printf("   Free or Isn  : %d\n", e.FreeOrIsn)
	}
	return nil
}

// RenameFile rename database file
func RenameFile(clientInstance *client.AdabasAdmin, dbid int, fnr int, newName string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewPutAdabasFileParameterParams()
	params.Dbid = float64(dbid)
	params.Name = &newName
	params.FileOperation = strconv.Itoa(fnr) + ":rename"
	resp, err := clientInstance.Online.PutAdabasFileParameter(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.PutAdabasFileParameterBadRequest:
			response := err.(*online.PutAdabasFileParameterBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println("Status: ", resp.Payload.Status.Message)
	return nil
}

// RenumberFile renumber database file
func RenumberFile(clientInstance *client.AdabasAdmin, dbid int, fnr int, newNumber string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewPutAdabasFileParameterParams()
	params.Dbid = float64(dbid)
	number, err := strconv.Atoi(newNumber)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	flNumber := float64(number)
	params.Number = &flNumber
	params.FileOperation = strconv.Itoa(fnr) + ":renumber"
	resp, err := clientInstance.Online.PutAdabasFileParameter(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.PutAdabasFileParameterBadRequest:
			response := err.(*online.PutAdabasFileParameterBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println("Status: ", resp.Payload.Status.Message)
	return nil
}

func loadFdt(fdt string) string {
	fmt.Println("Loading FDT file at " + fdt)
	raw, err := os.Open(fdt[4:])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	scanner := bufio.NewScanner(raw)
	var buffer bytes.Buffer
	r := regexp.MustCompile(" *;.*")
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, line)
		if line != "" {
			line = r.ReplaceAllString(line, "")
			if line != "" {
				if buffer.Len() > 0 {
					buffer.WriteString("%")
				}
				buffer.WriteString(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return buffer.String()
}

func createFileInstance(dbid int, fnr int, input InputList) *models.FduFdt {
	fdu := &models.FduFdt{}
	loadedFdt := ""
	fdu.FileNumber = int64(fnr)
	fdu.FduOptions = &models.FduFdtFduOptions{}
	for _, il := range input {
		if strings.HasPrefix(il, "fdt:") {
			loadedFdt = loadFdt(il)
		} else if strings.HasPrefix(il, "fdu:") {
			fileName := il[4:]
			fmt.Println("Loading FDU file at " + fileName)
			raw, err := ioutil.ReadFile(fileName)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := json.Unmarshal(raw, fdu); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("Prefix fdt: or fdu: missing")
		}
	}
	if fdu.FduOptions.FduName == "" {
		fmt.Println("FDU definition wrong, name missing")
		return nil
	}
	if loadedFdt != "" {
		fdu.FdtDefinition = &loadedFdt
	}
	return fdu
}

// CreateFile create database file
func CreateFile(clientInstance *client.AdabasAdmin, dbid int, fnr int, input InputList, auth runtime.ClientAuthInfoWriter) error {
	if len(input) == 0 {
		fmt.Println("Please add -input parameter for FDU and FDT")
		return fmt.Errorf("Please add -input parameter for FDU and FDT")
	}
	params := online.NewCreateAdabasFileParams()
	params.Dbid = float64(dbid)
	params.Fdufdt = createFileInstance(dbid, fnr, input)
	if params.Fdufdt == nil {
		return fmt.Errorf("Error parsing file")
	}
	resp, err := clientInstance.Online.CreateAdabasFile(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.CreateAdabasFileBadRequest:
			response := err.(*online.CreateAdabasFileBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err, reflect.TypeOf(err))
		}
		return err
	}
	fmt.Println("Status: ", resp.Payload.Status.Message)
	return nil
}

// DeleteFile delete database file
func DeleteFile(clientInstance *client.AdabasAdmin, dbid int, fnr int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewDeleteFileParams()
	params.Dbid = float64(dbid)
	params.FileOperation = float64(fnr)
	resp, err := clientInstance.OnlineOffline.DeleteFile(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.DeleteFileBadRequest:
			response := err.(*online_offline.DeleteFileBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Printf(" Adabas status deleting file: %s", resp.Payload.Status.Message)
	p.Println()
	return nil
}
