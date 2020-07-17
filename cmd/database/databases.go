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
	"time"

	"github.com/go-openapi/runtime"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"softwareag.com/client"
	"softwareag.com/client/offline"
	"softwareag.com/client/online"
	"softwareag.com/client/online_offline"
	"softwareag.com/models"
)

// List list databases
func List(clientInstance *client.AdabasAdmin, auth runtime.ClientAuthInfoWriter) error {
	// resp, err := client.Default.OnlineOffline.GetDatabases(nil, auth)
	resp, err := clientInstance.OnlineOffline.GetDatabases(nil, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabasesBadRequest:
			response := err.(*online_offline.GetDatabasesBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Printf(" %3s   %-16s    %8s    %s\n", "Dbid", "Name", "Active", "Version")
	fmt.Println()
	for _, d := range resp.Payload.Database {
		fmt.Printf("  %03d [%-16s]   %8v    %s\n", d.Dbid, d.Name, d.Active, d.Version)
	}
	fmt.Println()
	return nil
}

// Environment Operation init operations on database
func Environment(clientInstance *client.AdabasAdmin, auth runtime.ClientAuthInfoWriter) error {
	return nil
}

// Operation init operations on database
func Operation(clientInstance *client.AdabasAdmin, dbid int, operation string, auth runtime.ClientAuthInfoWriter) error {
	if operation != "" {
		fmt.Printf("\nSend following operation to database %v: %s\n", dbid, operation)
	} else {
		fmt.Printf("\nGet database information %v\n", dbid)
	}
	operation = strconv.Itoa(dbid) + ":" + operation
	params := online_offline.NewDatabaseOperationParams()
	params.DbidOperation = operation
	resp, accepted, err := clientInstance.OnlineOffline.DatabaseOperation(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.DatabaseOperationBadRequest:
			response := err.(*online_offline.DatabaseOperationBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	if accepted != nil {
		fmt.Printf("Database status dbid=%d %s\n", accepted.Payload.Status.Dbid, accepted.Payload.Status.Message)
	}
	if resp != nil {
		if resp.Payload.Database != nil {
			fmt.Printf("Database status dbid=%d %s\n", resp.Payload.Database.Dbid, resp.Payload.Database.Status)
		} else {
			fmt.Printf("Database operation inited successfully\n")
		}
	}
	return nil
}

func createDatabaseInstance(dbid int, input string) *models.Database {
	database := &models.Database{}
	if input == "" {
		database.Name = "DEMODB"
		database.LoadDemo = true
		database.Dbid = int64(dbid)
		container := &models.Container{BlockSize: "8K", ContainerSize: "60M",
			Path: fmt.Sprintf("${ADADATADIR}/db%03d/ASSO1.%03d", dbid, dbid)}
		database.ContainerList = append(database.ContainerList, container)
		container = &models.Container{BlockSize: "32K", ContainerSize: "20M",
			Path: fmt.Sprintf("${ADADATADIR}/db%03d/ASSO2.%03d", dbid, dbid)}
		database.ContainerList = append(database.ContainerList, container)
		container = &models.Container{BlockSize: "32K", ContainerSize: "100M",
			Path: fmt.Sprintf("${ADADATADIR}/db%03d/DATA1.%03d", dbid, dbid)}
		database.ContainerList = append(database.ContainerList, container)
		container = &models.Container{BlockSize: "4K", ContainerSize: "20M",
			Path: fmt.Sprintf("${ADADATADIR}/db%03d/WORK.%03d", dbid, dbid)}
		database.ContainerList = append(database.ContainerList, container)
		database.CheckpointFile = 1
		database.SecurityFile = 2
		database.UserFile = 3
		return database
	}
	raw, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := json.Unmarshal(raw, database); err != nil {
		log.Fatal(err)
	}
	return database
}

// Status  database online state
func Status(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewDatabaseOperationParams()
	params.DbidOperation = strconv.Itoa(dbid)
	resp, accepted, err := clientInstance.OnlineOffline.DatabaseOperation(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.DatabaseOperationBadRequest:
			response := err.(*online_offline.DatabaseOperationBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		case *online_offline.DatabaseOperationAccepted:
			response := err.(*online_offline.DatabaseOperationAccepted)
			fmt.Println(response)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Printf(" Adabas status of database %d: ", resp.Payload.Database.Dbid)
	if resp != nil {
		p.Printf(" %s\n", resp.Payload.Database.Status)
	}
	if accepted != nil {
		p.Printf(" %s\n", accepted.Payload.Status.Message)
	}
	p.Println()
	return nil
}

// Create create database
func Create(clientInstance *client.AdabasAdmin, dbid int, input string, auth runtime.ClientAuthInfoWriter) error {
	params := offline.NewPostAdabasDatabaseParams()
	params.Database = createDatabaseInstance(dbid, input)
	if dbid > 0 {
		params.Database.Dbid = int64(dbid)
	}
	resp, err := clientInstance.Offline.PostAdabasDatabase(params, auth)
	if err != nil {
		switch err.(type) {
		case *offline.PostAdabasDatabaseBadRequest:
			response := err.(*offline.PostAdabasDatabaseBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Println(" Adabas status of database creation:")
	p.Println()
	p.Printf(" %s", resp.Payload.Status.Message)
	return nil
}

// Delete database
func Delete(clientInstance *client.AdabasAdmin, dbid int, input string, auth runtime.ClientAuthInfoWriter) error {
	params := offline.NewDeleteAdabasDatabaseParams()
	params.DbidOperation = float64(dbid)
	resp, err := clientInstance.Offline.DeleteAdabasDatabase(params, auth)
	if err != nil {
		switch err.(type) {
		case *offline.DeleteAdabasDatabaseBadRequest:
			response := err.(*offline.DeleteAdabasDatabaseBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Printf(" Adabas status of database delete: %s", resp.Payload.Status.Message)
	p.Println()
	return nil
}

// Rename database
func Rename(clientInstance *client.AdabasAdmin, dbid int, param string, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewPutDatabaseResourceParams()
	params.DbidOperation = strconv.Itoa(dbid)
	params.Name = param
	resp, acpt, err := clientInstance.OnlineOffline.PutDatabaseResource(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.PutDatabaseResourceBadRequest:
			response := err.(*online_offline.PutDatabaseResourceBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	if resp != nil {
		if resp.Payload.Database != nil {
			p.Printf(" Adabas status of database delete: %s", resp.Payload.Database.Status)
		}
	}
	if acpt != nil {
		p.Printf(" Adabas status of database delete: %s", acpt.Payload.Status.Message)
	}
	p.Println()
	return nil
}

// NucleusLog show nucleus log
func NucleusLog(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseNucleusLogParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.OnlineOffline.GetDatabaseNucleusLog(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseNucleusLogBadRequest:
			response := err.(*online_offline.GetDatabaseNucleusLogBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Printf("\nDatabase %03d Nucleus log:\n", dbid)
	fmt.Println(resp.Payload.Log.Log)
	return nil
}

// Information database information
func Information(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseGcbParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
	params.Dbid = float64(dbid)
	resp, err := clientInstance.OnlineOffline.GetDatabaseGcb(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseGcbBadRequest:
			response := err.(*online_offline.GetDatabaseGcbBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Printf("Database %03d information:\n", dbid)
	p.Println()
	p.Printf("Dbid                : %d\n", resp.Payload.Gcb.Dbid)
	p.Printf("Name                : %s\n", resp.Payload.Gcb.Name)
	p.Printf("Version             : %s\n", resp.Payload.Gcb.StructureLevel)
	p.Printf("Architecture        : %s\n", resp.Payload.Gcb.Architecture)
	p.Printf("Created             : %s\n", time.Time(resp.Payload.Gcb.Date).Format("Mon Jan _2 15:04:05 2006"))
	p.Printf("Last changed        : %s\n", time.Time(resp.Payload.Gcb.TimeStampLog).Format("Mon Jan _2 15:04:05 2006"))
	p.Printf("PLOG count          : %d\n", resp.Payload.Gcb.PLOGCount)
	p.Printf("Current CLOG        : %d\n", resp.Payload.Gcb.CurrentCLOGNumber)
	p.Printf("Current PLOG        : %d\n", resp.Payload.Gcb.CurrentPLOGNumber)
	p.Printf("Flags               : %s\n", resp.Payload.Gcb.Flags)
	p.Printf("Maximum File Number : %d\n", resp.Payload.Gcb.MaxFileNumber)
	p.Printf("Files loaded        : %d\n", resp.Payload.Gcb.MaxFileNumberLoaded)
	p.Printf("Reserved Files\n")
	p.Printf(" Checkpoint File    : %d\n", resp.Payload.Gcb.CheckpointFile)
	p.Printf(" Security File      : %d\n", resp.Payload.Gcb.SecurityFile)
	p.Printf(" User File          : %d\n", resp.Payload.Gcb.ETDataFile)
	p.Printf("Replication\n")
	p.Printf(" Metadata File      : %d\n", resp.Payload.Gcb.ReplicationMetadataFile)
	p.Printf(" Command File       : %d\n", resp.Payload.Gcb.ReplicationCommandFile)
	p.Printf(" Transition File    : %d\n", resp.Payload.Gcb.ReplicationTransitionFile)
	p.Printf(" Timestamp Repl     : %s\n", time.Time(resp.Payload.Gcb.TimeStampReplication).Format("Mon Jan _2 15:04:05 2006"))
	p.Printf("Work\n")
	for i, e := range resp.Payload.Gcb.WORKExtents {
		if e.RABNunused != 0 {
			p.Printf(" Work extent        : %d\n", (i + 1))
			p.Printf("  Blocksize         : %v\n", e.BlockSize)
			p.Printf("  Device Type       : %v\n", e.DeviceType)
			p.Printf("  ID                : %v\n", e.ID)
			p.Printf("  Number            : %v\n", e.Number)
			p.Printf("  First RABN        : %v\n", e.RABNfirst)
			p.Printf("  Last RABN         : %v\n", e.RABNlast)
			p.Printf("  Unused RABN       : %v\n", e.RABNunused)
		}

	}
	p.Printf(" Work part 1        : %d\n", resp.Payload.Gcb.WORKPart1Size)
	return nil
}

// Activity database activity
func Activity(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseActStatsParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseActStats(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseActStatsBadRequest:
			response := err.(*online.GetDatabaseActStatsBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Println(" Adabas activity:")
	p.Println()
	p.Println(" I/O Activity                     Total   Throwbacks                       Total")
	p.Println(" ------------                     -----   ----------                       -----")
	p.Printf(" Buffer Pool               %12d   Waiting for UQ context    %12d\n", resp.Payload.Statistics.BufferPoolIO, resp.Payload.Statistics.ThbWaitUQContext)
	p.Printf(" WORK Read                 %12d   Waiting for ISN           %12d\n", resp.Payload.Statistics.WorkReads, resp.Payload.Statistics.ThbWaitIsn)
	p.Printf(" WORK Write                %12d   ET Sync                   %12d\n", resp.Payload.Statistics.WorkWrites, resp.Payload.Statistics.ThbEtSync)
	p.Printf(" PLOG Write                %12d   DWP Overflow              %12d\n", resp.Payload.Statistics.PlogWrites, resp.Payload.Statistics.ThbDWPOverflow)
	p.Printf(" NUCTMP                    %12d\n", -1)
	p.Printf(" NUCSRT                    %12d\n", -1)
	p.Println()
	p.Println(" Pool Hit Rate                    Total   Interrupts       Current         Total")
	p.Println(" -------------                    -----   ----------       -------         -----")
	p.Printf(" Buffer Pool                        %.1f%% WP Space Wait %10d    %10d\n", float64(resp.Payload.Statistics.BPHitRate), resp.Payload.Statistics.WPSpaceWaitCurrent, resp.Payload.Statistics.WpSpaceWaitTotal)
	p.Printf(" Format pool                        %.1f%%\n", float64(resp.Payload.Statistics.FPHitRate))
	return nil
}

// ThreadTable display thread table
func ThreadTable(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseThreadTableParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseThreadTable(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseThreadTableBadRequest:
			response := err.(*online.GetDatabaseThreadTableBadRequest)
			fmt.Printf(" Adabas thread table:\n\n")
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Println(" Adabas thread table:")
	p.Println()
	p.Println(" No     Cmd Count  File  Cmd  Status")
	p.Println(" --     ---------  ----  ---  ------")
	for _, t := range resp.Payload.Threads {
		p.Printf(" %2d    %10d %5d   %2s  %s\n", t.Thread, t.CommandCount, t.File, t.CommandCode, t.Status)
	}
	return nil
}

// Parameter show parameter
func Parameter(clientInstance *client.AdabasAdmin, dbid int, para string, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseParameterParams()
	params.Dbid = float64(dbid)
	if para == "" {
		fmt.Println("Please provide parameter static or dynamic to reference corresponding parameter configuration.")
		return fmt.Errorf("Please provide parameter static or dynamic to reference corresponding parameter configuration")
	}

	params.Type = strings.ToLower(para)
	if params.Type != "static" && params.Type != "dynamic" {
		fmt.Println("Parameter type must be static or dynamic to reference corresponding parameter configuration.")
		return fmt.Errorf("Parameter type must be static or dynamic to reference corresponding parameter configuration")
	}

	resp, err := clientInstance.OnlineOffline.GetDatabaseParameter(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseParameterBadRequest:
			response := err.(*online_offline.GetDatabaseParameterBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Printf(" Adabas %s parameter info:\n", para)
	dbParameter := resp.Payload.Parameter
	val := reflect.ValueOf(*dbParameter)
	typ := reflect.TypeOf(*dbParameter)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		n, ok := field.Tag.Lookup("json")
		if !ok {
			n = field.Name
		} else {
			ec := strings.IndexByte(n, ',')
			if ec > 0 {
				n = n[:ec]
			}
		}
		v := val.Field(i)
		fmt.Println("    ", n, "=", v)
	}
	return nil
}

// ParameterInfo show parameter info
func ParameterInfo(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseParameterInfoParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.OnlineOffline.GetDatabaseParameterInfo(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseParameterInfoBadRequest:
			response := err.(*online_offline.GetDatabaseParameterInfoBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println()
	p.Printf(" Adabas parameter info:\n")
	for _, parameter := range resp.Payload.ParameterInfo.Parameter {
		if parameter.Acronym != "" {
			p.Printf("[%s]\n", parameter.Acronym)
			p.Printf("%-20s: %s\n", parameter.Name, parameter.Description)
			p.Printf("  Dynamic: %v\n", parameter.IsDynamic)
			switch parameter.Acronym {
			case "AR":
				d := checkArConflict(parameter.DefaultValue)
				c := checkArConflict(parameter.InifileValue)
				p.Printf("  Default: %14s  Configuration: %14s\n", d, c)
			case "OPTIONS":
				p.Printf("  Default: %s\n", parameter.DefaultValue)
				//	c := checkOptions(parameter.InifileValue)
				//	fmt.Println("XXXXX ", c, parameter.InifileValue)
				p.Printf("  Configuration: %s\n", parameter.InifileValue)
				//	o := checkOptions(parameter.OnlineValue)
				p.Printf("  Online:        %s\n", parameter.OnlineValue)
			case "USEREXITS":
				p.Printf("  Default:\n")
				c := checkUserexits(parameter.InifileValue)
				p.Printf("  Configuration: %s\n", c)
				o := checkUserexits(parameter.OnlineValue)
				p.Printf("  Online:        %s\n", o)
			case "LOGGING":
				p.Printf("  Default:\n")
				c := checkLogging(parameter.InifileValue)
				p.Printf("  Configuration: %s\n", c)
				o := checkLogging(parameter.OnlineValue)
				p.Printf("  Online:        %s\n", o)
			default:
				p.Printf("  Default: %14s  Configuration: %14s  Online: %14s\n", parameter.DefaultValue, parameter.InifileValue, parameter.OnlineValue)
			}
			if parameter.IsMinValueAvailable {
				p.Printf("  Minimum: %14d\n", parameter.MinValue)
			}
			if *parameter.IsMaxValueAvailable {
				p.Printf("  Maximum: %14d\n", parameter.MaxValue)
			}
			p.Println()
		}
	}
	p.Println()
	return nil
}

type parameterSet struct {
	name  string
	value string
}

// SetParameter set parameter
func SetParameter(clientInstance *client.AdabasAdmin, dbid int, param string, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewPutAdabasParameterParams()
	params.Dbid = float64(dbid)
	pmap := make(map[string]string)
	params.Type = "static"
	options := ""
	for _, p := range strings.Split(param, ",") {
		fmt.Println("Got", p)
		if options != "" {
			if strings.Contains(p, ")") {
				options = options + "," + strings.Replace(p, ")", "", 1)
				sendOption := options
				params.OPTIONS = &sendOption
				fmt.Println("End options=" + options)
				options = ""
				fmt.Printf("Options=%s", *params.OPTIONS)

			} else {
				options = options + "," + p
				fmt.Println("Add options=" + options)
			}
			continue
		}
		v := strings.Split(p, "=")
		fmt.Println(v)
		if len(v) != 2 {
			fmt.Printf("Parameter %s not valid, need of the type: param1=1,param2=ON,param3=(A,B)", param)
			return fmt.Errorf("Parameter %s not valid, need of the type: param1=1,param2=ON,param3=(A,B)", param)
		}
		if strings.ToLower(v[0]) == "type" {
			if strings.ToLower(v[1]) == "dynamic" {
				params.Type = "dynamic"
			}
		} else {
			if v[0] == "OPTIONS" {
				options = strings.Replace(v[1], "(", "", 1)
				if strings.Contains(v[1], ")") {
					options = strings.Replace(options, ")", "", 1)
					sendOption := options
					if sendOption == "" {
						sendOption = " "
					}
					params.OPTIONS = &sendOption
					fmt.Println("End options=" + *params.OPTIONS)
					options = ""
				} else {
					fmt.Println("Found options=" + options)
				}
			} else {
				n := strings.Replace(v[0], "_", "", -1)
				pmap[n] = v[1]
				vp := reflect.ValueOf(params).Elem()
				f := vp.FieldByName(n)
				if f.IsValid() {
					if f.IsNil() && f.CanSet() {
						switch n {
						case "NT", "TT", "NU", "NCL", "NISNHQ", "TNAE", "TNAA", "TNAX", "LAB", "LABX", "LBP",
							"LWP", "LPXA", "ADATCPPORT", "ADATCPRECEIVER", "ADATCPATB", "ADATCPCONNECTIONS", "SSLVERIFY",
							"APUUNITS", "APURECVS", "APUWORKERS", "SSLPORT",
							"RPLBLOCKS", "RPLTOTAL", "RPLRECORDS", "WRITELIMIT":
							i, err := strconv.Atoi(v[1])
							if err != nil {
								fmt.Println("Incorrect value for " + v[0])
								return fmt.Errorf("Incorrect value for " + v[0])
							}
							i64 := int64(i)
							f.Set(reflect.ValueOf(&i64))
						case "ADATCP", "PLOG", "BI":
							var i bool
							switch strings.ToLower(v[1]) {
							case "on", "yes", "true":
								i = true
							default:
								i = false
							}
							f.Set(reflect.ValueOf(&i))

						default:
							f.Set(reflect.ValueOf(&v[1]))
						}

					}
				}
				fmt.Println("Set", f, f.Type())
			}
		}
	}

	resp, err := clientInstance.OnlineOffline.PutAdabasParameter(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.PutAdabasParameterBadRequest:
			response := err.(*online_offline.PutAdabasParameterBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf(" Adabas parameter: %s", resp.Payload.Status.Message)
	fmt.Println()
	return nil
}

func checkArConflict(value string) string {
	if value == "1" {
		return "CONTINUE"
	}
	return "ABORT"
}

const (
	aifSpaOptionsTruncation      = 1   /**< Truncation */
	aifSpaOptionsUtilitiesOnly   = 2   /**< Utilities only */
	aifSpaOptionsLocalUtilities  = 4   /**< Local utilities */
	aifSpaOptionsOpenRequired    = 8   /**< Open required */
	aifSpaOptionsFaultTolerantAR = 16  /**< Fault tolerant auto restart */
	aifSpaOptionsAutorestartOnly = 32  /**< Autorestart only */
	aifSpaOptionsReadOnly        = 64  /**< Read only database */
	aifSpaOptionsXA              = 128 /**< XA support enabled */
	aifSpaOptionsAutoExpand      = 256 /**< Auto expand */
	aifSpaOptionsDeactivate      = 512 /**< Deactivate dynamic options  */
)

func checkOptions(opt string) string {
	value, _ := strconv.Atoi(opt)
	buffer := bytes.Buffer{}
	for i := uint(0); i < 9; i++ {
		if value&(1<<i) != 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			switch 1 << i {
			case aifSpaOptionsTruncation:
				buffer.WriteString("TRUNCATION")
			case aifSpaOptionsUtilitiesOnly:
				buffer.WriteString("UTILITIES_ONLY")
			case aifSpaOptionsLocalUtilities:
				buffer.WriteString("LOCAL_UTILITIES")
			case aifSpaOptionsOpenRequired:
				buffer.WriteString("OPEN_REQUIRED")
			case aifSpaOptionsFaultTolerantAR:
				buffer.WriteString("FAULT_TOLERANT_AR")
			case aifSpaOptionsAutorestartOnly:
				buffer.WriteString("AUTORESTART_ONLY")
			case aifSpaOptionsReadOnly:
				buffer.WriteString("READ_ONLY")
			case aifSpaOptionsXA:
				buffer.WriteString("XA")
			case aifSpaOptionsAutoExpand:
				buffer.WriteString("AUTO_EXPAND")
			case aifSpaOptionsDeactivate:
				buffer.WriteString("DEACTIVATE")
			default:
				buffer.WriteString("XXXXX-" + strconv.Itoa(int(i)))
			}
		}
	}
	return buffer.String()
}

const (
	aifSpaUserexit01 = 1
	aifSpaUserexit02 = 2
	aifSpaUserexit04 = (1 << 3)
	aifSpaUserexit11 = (1 << 10)
	aifSpaUserexit14 = (1 << 13)
)

func checkUserexits(ue string) string {
	value, _ := strconv.Atoi(ue)
	buffer := bytes.Buffer{}
	for i := uint(0); i < 14; i++ {
		if value&(1<<i) != 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			switch 1 << i {
			case aifSpaUserexit01:
				buffer.WriteString("1")
			case aifSpaUserexit02:
				buffer.WriteString("2")
			case aifSpaUserexit04:
				buffer.WriteString("4")
			case aifSpaUserexit11:
				buffer.WriteString("11")
			case aifSpaUserexit14:
				buffer.WriteString("14")
			default:
				buffer.WriteString("XXXXX-" + strconv.Itoa(int(i)))
			}
		}
	}
	return buffer.String()
}

const (
	aifSpaLoggingCB      = (1 << 0)  /**< Control Block */
	aifSpaLoggingFB      = (1 << 1)  /**< Format Buffer */
	aifSpaLoggingIB      = (1 << 2)  /**< ISN Buffer */
	aifSpaLoggingIO      = (1 << 3)  /**< IO Activity */
	aifSpaLoggingRB      = (1 << 4)  /**< Record Buffer */
	aifSpaLoggingSB      = (1 << 5)  /**< Search Buffer */
	aifSpaLoggingVB      = (1 << 6)  /**< Value Buffer */
	aifSpaLoggingOFF     = (1 << 7)  /**< Logging disabled */
	aifSpaLoggingBD      = (1 << 8)  /**< Buffer Description */
	aifSpaLoggingENABLED = (1 << 9)  /**< Logging enabled  */
	aifSpaLoggingAR      = (1 << 10) /**< Fault tolerant Auto Restart */
)

func checkLogging(logging string) string {
	value, _ := strconv.Atoi(logging)
	buffer := bytes.Buffer{}
	for i := uint(0); i < 9; i++ {
		if value&(1<<i) != 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(",")
			}
			switch 1 << i {
			case aifSpaLoggingCB:
				buffer.WriteString("CB")
			case aifSpaLoggingFB:
				buffer.WriteString("FB")
			case aifSpaLoggingIB:
				buffer.WriteString("IB")
			case aifSpaLoggingIO:
				buffer.WriteString("IO")
			case aifSpaLoggingRB:
				buffer.WriteString("RB")
			case aifSpaLoggingSB:
				buffer.WriteString("SB")
			case aifSpaLoggingVB:
				buffer.WriteString("VB")
			case aifSpaLoggingOFF:
				buffer.WriteString("OFF")
			case aifSpaLoggingBD:
				buffer.WriteString("BD")
			case aifSpaLoggingENABLED:
				buffer.WriteString("ENABLED")
			case aifSpaLoggingAR:
				buffer.WriteString("AR")
			default:
				buffer.WriteString("XXXXX-" + strconv.Itoa(int(i)))
			}
		}
	}
	return buffer.String()
}

// Container list database container
func Container(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetDatabaseContainerParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.OnlineOffline.GetDatabaseContainer(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetDatabaseContainerBadRequest:
			response := err.(*online_offline.GetDatabaseContainerBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Printf("Database %03d container:\n", dbid)
	p.Println()
	for _, c := range resp.Payload.Container.ContainerList {
		p.Printf(" %5s%-2d %s %8d%s %8d%s  %8d:%8d %6d  %s\n", c.Type, c.ContainerNumber, c.DeviceType,
			c.BlockSize, c.BlockUnit, c.Size, c.SizeUnit, c.FirstExtentRabn, c.LastExtentRabn,
			c.FirstUnusedRabn, c.Path)
	}
	p.Println()
	p.Printf("Database %03d free space table:\n", dbid)
	p.Println()
	for _, c := range resp.Payload.Container.FreeSpaceTable {
		p.Printf(" %5s %10d %10d %4d\n", c.Type, c.FirstRABN, c.LastRABN, c.BlockSize)
	}
	return nil
}

// Checkpoints list checkpoints in an specific range
func Checkpoints(clientInstance *client.AdabasAdmin, dbid int, timeRange string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseCheckpointsParams()
	params.Dbid = float64(dbid)
	if timeRange == "" {
		fmt.Println("Query checkpoint of the last 24 hours")

		t := time.Now()
		layout := "2006-01-02 15:04:05"
		t1 := t.AddDate(0, 0, -1)
		start := t1.Format(layout)
		end := t.Format(layout)
		params.StartTime = &start
		params.EndTime = &end
	} else {
		rg := regexp.MustCompile("_")
		tr := rg.ReplaceAllString(timeRange, " ")
		r := strings.Split(tr, ",")
		start := r[0]
		end := r[1]
		params.StartTime = &start
		params.EndTime = &end
	}
	resp, err := clientInstance.Online.GetDatabaseCheckpoints(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseCheckpointsBadRequest:
			response := err.(*online.GetDatabaseCheckpointsBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println("\nQuery checkpoint from ", *params.StartTime, " to ", *params.EndTime)
	for _, c := range resp.Payload.Checkpoints {
		fmt.Println(c.Name, c.Session, c.Date, c.Details)
	}
	return nil
}

// DeleteCheckpoints delete checkpoints in an specific range
func DeleteCheckpoints(clientInstance *client.AdabasAdmin, dbid int, timeRange string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewDeleteDatabaseCheckpointsParams()
	params.Dbid = float64(dbid)
	if timeRange == "" {
		fmt.Println("Please provide time range of the checkpoints which should be delete. Example:")
		fmt.Println("<admincmd> -param -param '2018-05-15_01:00:00,2018-05-20_00:00:00'")
		return fmt.Errorf("Checkpoint range parameter missing")
	}
	rg := regexp.MustCompile("_")
	tr := rg.ReplaceAllString(timeRange, " ")
	r := strings.Split(tr, ",")
	start := r[0]
	end := r[1]
	params.StartTime = &start
	params.EndTime = &end

	fmt.Println("Query checkpoint from ", *params.StartTime, " to ", *params.EndTime)
	resp, err := clientInstance.Online.DeleteDatabaseCheckpoints(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.DeleteDatabaseCheckpointsBadRequest:
			response := err.(*online.DeleteDatabaseCheckpointsBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Println()
	fmt.Printf(" Adabas status of delete checkpoint in range of %s to %s: %s",
		*params.StartTime, *params.EndTime, resp.Payload.Status.Message)
	fmt.Println()
	return nil
}

// Ucb list UCBs
func Ucb(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetUCBParams()
	params.Dbid = float64(dbid)

	resp, err := clientInstance.OnlineOffline.GetUCB(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetUCBBadRequest:
			response := err.(*online_offline.GetUCBBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Println(" UCB entries:")
	fmt.Println()
	fmt.Printf(" %-20s %-10s %-8s %-8s %-8s\n", "Date/Time", "Entry ID", "Utility", "Mode", "Files")
	for _, c := range resp.Payload.UCB.UCB {
		s, _ := json.Marshal(c.UcbFiles)
		fmt.Printf(" %-20s %-10d %-8s %-8s %s\n", c.Date, c.Sequence, c.ID, c.DBMode, s)
	}
	return nil
}

// DeleteUcb delete UCB entry
func DeleteUcb(clientInstance *client.AdabasAdmin, dbid int, param string, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewDeleteUCBParams()
	params.Dbid = float64(dbid)
	ucbid, err := strconv.Atoi(param)
	if err != nil {
		fmt.Println("Error UCB id parameter not numeric: ", err)
		return err
	}
	params.Ucbid = int64(ucbid)

	resp, respErr := clientInstance.OnlineOffline.DeleteUCB(params, auth)
	err = respErr
	if err != nil {
		switch err.(type) {
		case *online_offline.DeleteUCBBadRequest:
			response := err.(*online_offline.DeleteUCBBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Println()
	fmt.Printf(" Adabas status of UCB delete: %s", resp.Payload.Status.Message)
	fmt.Println()
	return nil
}
