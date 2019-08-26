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
package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"golang.org/x/crypto/ssh/terminal"
	"softwareag.com/client"
	"softwareag.com/client/environment"
	"softwareag.com/cmd/database"
	"softwareag.com/cmd/filebrowser"
	"softwareag.com/cmd/job"
)

type display int

const (
	unknown display = iota
	env
	list
	start
	shutdown
	cancel
	abort
	info
	userqueue
	cmdqueue
	holdqueue
	highwater
	commandstats
	bp
	activity
	threadtable
	createdatabase
	deletedatabase
	renamedatabase
	parameter
	parameterinfo
	setparameter
	nucleuslog
	files
	file
	deletefile
	renumberfile
	refreshfile
	fields
	information
	container
	renamefile
	createfile
	checkpoints
	joblist
	jobstart
	deletejob
	deletejobexec
	createjob
	joblog
	listucb
	deleteucb
	addfields
	status
	filelocations
	listfiles
	downloadfile
	uploadfile
)

const (
	adabasAdminPassword = "ADABAS_ADMIN_PASSWORD"
	adabasAdminURL      = "ADABAS_ADMIN_URL"
)

type displayInfo struct {
	id             display
	cmdShort       string
	cmdDescription string
}

var aborted = false

var displayName = []displayInfo{
	displayInfo{id: unknown, cmdShort: "unknown"},
	displayInfo{id: env, cmdShort: "env", cmdDescription: "List Adabas environment version"},
	displayInfo{id: list, cmdShort: "list", cmdDescription: "List all Adabas databases"},
	displayInfo{id: start, cmdShort: "start", cmdDescription: "Start Adabas database"},
	displayInfo{id: shutdown, cmdShort: "shutdown", cmdDescription: "Shutdown Adabas database"},
	displayInfo{id: cancel, cmdShort: "cancel", cmdDescription: "Cancel Adabas database"},
	displayInfo{id: abort, cmdShort: "abort", cmdDescription: "Abort Adabas database"},
	displayInfo{id: info, cmdShort: "info", cmdDescription: "Retrieve Adabas database information"},
	displayInfo{id: userqueue, cmdShort: "userqueue", cmdDescription: "Display current user queue"},
	displayInfo{id: cmdqueue, cmdShort: "cmdqueue", cmdDescription: "Display current command queue"},
	displayInfo{id: holdqueue, cmdShort: "holdqueue", cmdDescription: "Display current hold queue"},
	displayInfo{id: highwater, cmdShort: "highwater", cmdDescription: "Display high water mark"},
	displayInfo{id: commandstats, cmdShort: "commandstats", cmdDescription: "Display Adabas command statistics"},
	displayInfo{id: bp, cmdShort: "bp", cmdDescription: "Display Adabas buffer pool statistics"},
	displayInfo{id: activity, cmdShort: "activity", cmdDescription: "Display Adabas activity"},
	displayInfo{id: threadtable, cmdShort: "threadtable", cmdDescription: "Display Adabas thread table"},
	displayInfo{id: createdatabase, cmdShort: "createdatabase", cmdDescription: "Create new Adabas database"},
	displayInfo{id: deletedatabase, cmdShort: "deletedatabase", cmdDescription: "Delete a Adabas database"},
	displayInfo{id: renamedatabase, cmdShort: "renamedatabase", cmdDescription: "Rename a Adabas database"},
	displayInfo{id: parameter, cmdShort: "parameter", cmdDescription: "List database parameter information"},
	displayInfo{id: parameterinfo, cmdShort: "parameterinfo", cmdDescription: "List database parameter information with minimum and maximum ranges"},
	displayInfo{id: setparameter, cmdShort: "setparameter", cmdDescription: "Set database parameter"},
	displayInfo{id: nucleuslog, cmdShort: "nucleuslog", cmdDescription: "Display Adabas nucleus log"},
	displayInfo{id: files, cmdShort: "files", cmdDescription: "Display Adabas file list"},
	displayInfo{id: file, cmdShort: "file", cmdDescription: "Display Adabas file"},
	displayInfo{id: deletefile, cmdShort: "deletefile", cmdDescription: "Delete Adabas file"},
	displayInfo{id: renumberfile, cmdShort: "renumberfile", cmdDescription: "Renumber Adabas file"},
	displayInfo{id: refreshfile, cmdShort: "refreshfile", cmdDescription: "Refresh Adabas file"},
	displayInfo{id: fields, cmdShort: "fields", cmdDescription: "Display Adabas file definition table"},
	displayInfo{id: information, cmdShort: "information", cmdDescription: "Display Adabas database information"},
	displayInfo{id: container, cmdShort: "container", cmdDescription: "Display Adabas database container"},
	displayInfo{id: renamefile, cmdShort: "renamefile", cmdDescription: "Rename Database file"},
	displayInfo{id: createfile, cmdShort: "createfile", cmdDescription: "Create Database file"},
	displayInfo{id: checkpoints, cmdShort: "checkpoints", cmdDescription: "Display Database checkpoints. Without parameter it shows one day.\n  Parameter example with from and to parameter: 2018-05-15_01:00:00,2018-05-20_00:00:00"},
	displayInfo{id: joblist, cmdShort: "joblist", cmdDescription: "Job control list"},
	displayInfo{id: jobstart, cmdShort: "jobstart", cmdDescription: "Start a specific job"},
	displayInfo{id: deletejob, cmdShort: "deletejob", cmdDescription: "Delete a specific job and the execution log"},
	displayInfo{id: deletejobexec, cmdShort: "deletejobexec", cmdDescription: "Delete the execution log of a job"},
	displayInfo{id: createjob, cmdShort: "createjob", cmdDescription: "Create a new specific job"},
	displayInfo{id: joblog, cmdShort: "joblog", cmdDescription: "Job entry log"},
	displayInfo{id: listucb, cmdShort: "listucb", cmdDescription: "List Adabas UCB entries"},
	displayInfo{id: deleteucb, cmdShort: "deleteucb", cmdDescription: "Delete Adabas UCB entry"},
	displayInfo{id: addfields, cmdShort: "addfields", cmdDescription: "Add Adabas fields"},
	displayInfo{id: status, cmdShort: "status", cmdDescription: "Adabas database online state"},
	displayInfo{id: filelocations, cmdShort: "filelocations", cmdDescription: "List all available file locations"},
	displayInfo{id: listfiles, cmdShort: "listfiles", cmdDescription: "List file in file location"},
	displayInfo{id: downloadfile, cmdShort: "downloadfile", cmdDescription: "Download file out of file location"},
	displayInfo{id: uploadfile, cmdShort: "uploadfile", cmdDescription: "Upload file to file location"}}

func displayValue(name string) display {
	for i := 0; i < len(displayName); i++ {
		if name == displayName[i].cmdShort {
			return display(i)
		}
	}
	return unknown
}

/* Output of usage, displays flags and commands possible */
func usage() {
	flag.Usage()
	fmt.Println("\nPossible commands:")
	for _, d := range displayName {
		if d.id != unknown {
			fmt.Printf(" %s\n   \t%s\n", d.cmdShort, d.cmdDescription)
		}
	}

}

func main() {
	var restURL string
	var input database.InputList

	user := flag.String("user", "admin", "User name of the main administrator (default: admin)")
	passwd := flag.String("passwd", "", "Password of administration, may be predefined using environment variable ADABAS_ADMIN_PASSWORD")
	dbid := flag.Int("dbid", 0, "Adabas Database id")
	fnr := flag.Int("fnr", 0, "Adabas Database file")
	sleep := flag.Int("repeat", 0, "Repeat display after given seconds")
	ignoreTLS := flag.Bool("ignoreTLS", false, "Ignore TLS certificate validation")
	param := flag.String("param", "", "Method specific parameters")

	flag.Var(&input, "input", "Input configuration")
	flag.StringVar(&restURL, "url", "", "Remote RESTful server location URL, may be predefined using environment variable ADABAS_ADMIN_URL (example: localhost:8120, https://localhost:8121)")
	flag.Parse()

	// Check URL location is set
	if restURL == "" {
		restURL = os.Getenv(adabasAdminURL)
		if restURL == "" {
			fmt.Println("No host URL provided, use -url parameter or environment setting in " + adabasAdminURL)
			usage()
			os.Exit(1)
		}
	}

	// Get additional flags
	args := flag.Args()
	// Ask for user and password
	username := *user
	password := *passwd

	printStart(restURL, username)

	for _, a := range args {
		if displayValue(a) == unknown {
			fmt.Printf("Unknown command: %s\n", a)
			usage()
			aborted = true
			return
		}
	}

	if password == "" {
		password = os.Getenv(adabasAdminPassword)
		if password == "" {
			password = credentials()
		}
	}
	cookieJar, _ := cookiejar.New(nil)
	ru := restURL
	if strings.HasPrefix(ru, "http") {
		ru = ru[strings.Index(ru, "://")+3:]
	}
	h, _, errx := net.SplitHostPort(ru)
	if errx != nil {
		fmt.Printf("Host url error %s: %v", restURL, errx)
		os.Exit(2)
	}
	cookieURL := &url.URL{Scheme: "http", Host: h, Path: "/adabas"}
	var cookie *http.Cookie
	auth := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, _ strfmt.Registry) error {
		//	if cookie == nil {
		cookies := cookieJar.Cookies(cookieURL)
		for _, c := range cookies {
			if c.Name == "ADAADMIN" {
				expiration := time.Now().Add(5 * time.Minute)
				cookie = &http.Cookie{Name: "ADAADMIN", Value: c.Value, Expires: expiration}
				r.SetHeaderParam("Cookie", cookie.String())
				break
			}
		}
		//	}
		encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		return r.SetHeaderParam("Authorization", "Basic "+encoded)
	})

	var transport *httptransport.Runtime
	if strings.HasPrefix(restURL, "http") {
		if strings.HasPrefix(restURL, "https") {
			restURL = restURL[8:]
			// create the transport
			transport = httptransport.New(restURL, "", []string{"https"})
			if *ignoreTLS {
				transport.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			}
		} else {
			restURL = restURL[7:]
			// create the transport
			transport = httptransport.New(restURL, "", []string{"http"})
		}

	} else {
		// create the transport
		transport = httptransport.New(restURL, "", []string{"http"})
	}

	// transport.EnableConnectionReuse()
	//	transport.Transport = &MyTransport{apiKey: "ADAADMIN", rt: transport.Transport}
	transport.Jar = cookieJar

	// create the API client, with the transport
	clientInstance := client.New(transport, strfmt.Default)
	if len(args) == 0 {
		version(clientInstance)
		return
	}

	// expiration := time.Now().Add(5 * time.Minute)
	// cookie := http.Cookie{Name: "myCookie", Value: "Hello World", Expires: expiration}
	// http.SetCookie(clientInstance, &cookie)

	defer printEnd(time.Now())

	for {
		var err error
		for _, a := range args {
			d := displayValue(a)
			switch d {
			case unknown:
				fmt.Printf("Unknown command: %s\n", a)
				usage()
				aborted = true
				os.Exit(4)
			case env:
				err = database.Environment(clientInstance, auth)
			case list:
				err = database.List(clientInstance, auth)
			case start:
				err = database.Operation(clientInstance, *dbid, "start", auth)
			case shutdown:
				err = database.Operation(clientInstance, *dbid, "shutdown", auth)
			case cancel:
				err = database.Operation(clientInstance, *dbid, "cancel", auth)
			case abort:
				err = database.Operation(clientInstance, *dbid, "abort", auth)
			case info:
				err = database.Operation(clientInstance, *dbid, "", auth)
			case userqueue:
				err = database.UserQueue(clientInstance, *dbid, auth)
			case cmdqueue:
				err = database.CommandQueue(clientInstance, *dbid, auth)
			case holdqueue:
				err = database.HoldQueue(clientInstance, *dbid, auth)
			case highwater:
				err = database.Highwater(clientInstance, *dbid, auth)
			case commandstats:
				err = database.CommandStats(clientInstance, *dbid, auth)
			case bp:
				err = database.BufferpoolStats(clientInstance, *dbid, auth)
			case activity:
				err = database.Activity(clientInstance, *dbid, auth)
			case threadtable:
				err = database.ThreadTable(clientInstance, *dbid, auth)
			case createdatabase:
				err = database.Create(clientInstance, *dbid, input.String(), auth)
			case deletedatabase:
				err = database.Delete(clientInstance, *dbid, input.String(), auth)
			case renamedatabase:
				err = database.Rename(clientInstance, *dbid, *param, auth)
			case parameter:
				err = database.Parameter(clientInstance, *dbid, *param, auth)
			case parameterinfo:
				err = database.ParameterInfo(clientInstance, *dbid, auth)
			case setparameter:
				err = database.SetParameter(clientInstance, *dbid, *param, auth)
			case nucleuslog:
				err = database.NucleusLog(clientInstance, *dbid, auth)
			case files:
				err = database.Files(clientInstance, *dbid, auth)
			case file:
				err = database.File(clientInstance, *dbid, *fnr, *param, auth)
			case deletefile:
				err = database.DeleteFile(clientInstance, *dbid, *fnr, auth)
			case renumberfile:
				err = database.RenumberFile(clientInstance, *dbid, *fnr, *param, auth)
			case refreshfile:
				err = database.RefreshFile(clientInstance, *dbid, *fnr, auth)
			case information:
				err = database.Information(clientInstance, *dbid, auth)
			case fields:
				err = database.Fields(clientInstance, *dbid, *fnr, auth)
			case container:
				err = database.Container(clientInstance, *dbid, auth)
			case renamefile:
				err = database.RenameFile(clientInstance, *dbid, *fnr, *param, auth)
			case createfile:
				err = database.CreateFile(clientInstance, *dbid, *fnr, input, auth)
			case checkpoints:
				err = database.Checkpoints(clientInstance, *dbid, *param, auth)
			case joblist:
				err = job.List(clientInstance, auth)
			case jobstart:
				err = job.Start(clientInstance, *param, auth)
			case deletejob:
				err = job.Delete(clientInstance, *param, auth)
			case deletejobexec:
				err = job.DeleteExecution(clientInstance, *param, auth)
			case createjob:
				err = job.Create(clientInstance, input.String(), auth)
			case joblog:
				err = job.Log(clientInstance, *param, auth)
			case listucb:
				err = database.Ucb(clientInstance, *dbid, auth)
			case deleteucb:
				err = database.DeleteUcb(clientInstance, *dbid, *param, auth)
			case addfields:
				err = database.AddFields(clientInstance, *dbid, *fnr, *param, auth)
			case status:
				err = database.Status(clientInstance, *dbid, auth)
			case filelocations:
				err = filebrowser.Locations(clientInstance, auth)
			case listfiles:
				err = filebrowser.List(clientInstance, *param, auth)
			case downloadfile:
				err = filebrowser.Download(clientInstance, *param, input.String(), auth)
			case uploadfile:
				err = filebrowser.Upload(clientInstance, *param, input.String(), auth)
			default:
				err = version(clientInstance)
			}
			if err != nil {
				break
			}
		}
		if err != nil {
			os.Exit(10)
		}

		if *sleep == 0 {
			break
		} else {
			time.Sleep(time.Duration(*sleep) * time.Second)
		}
	}
}

func printStart(location string, username string) {
	out := "2006/01/02 15:04:05"

	fmt.Println(time.Now().Format(out), "Adabas Administration RESTful client started")
	fmt.Println()
	fmt.Println(time.Now().Format(out), "Server: "+location)
	fmt.Println(time.Now().Format(out), "User:   "+username)
	fmt.Println()
}

func printEnd(start time.Time) {

	out := "2006/01/02 15:04:05"

	elapsed := time.Since(start)

	fmt.Println()
	if !aborted {
		fmt.Printf("%s Adabas Administration RESTful client took %s terminated\n", time.Now().Format(out), elapsed)
	} else {
		fmt.Printf("%s Adabas Administration RESTful client took %s aborted\n", time.Now().Format(out), elapsed)
	}
}

func version(clientInstance *client.AdabasAdmin) error {
	params := environment.NewGetVersionParams()
	resp, err := clientInstance.Environment.GetVersion(params)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Printf("Version %s %s\n", resp.Payload.Version, resp.Payload.Product)
	fmt.Printf("\nHandlers:\n")
	for _, h := range resp.Payload.Handler {
		fmt.Printf(" %s: %s\n", h.Name, h.Version)
	}
	return nil
}

func adabasEnv(clientInstance *client.AdabasAdmin, auth runtime.ClientAuthInfoWriter) {
	params := environment.NewGetEnvironmentsParams()
	resp, err := clientInstance.Environment.GetEnvironments(params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Print(resp.Payload.Environment)
}

func credentials() string {
	// reader := bufio.NewReader(os.Stdin)

	// fmt.Print("Enter Username: ")
	// username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error entering password:", err)
		os.Exit(2)
	}
	fmt.Println()
	password := string(bytePassword)

	// return strings.TrimSpace(username), strings.TrimSpace(password)
	return strings.TrimSpace(password)
}
