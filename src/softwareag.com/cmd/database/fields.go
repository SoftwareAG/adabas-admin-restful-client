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
	"fmt"

	"github.com/go-openapi/runtime"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"softwareag.com/client"
	"softwareag.com/client/online_offline"
	"softwareag.com/models"
)

// Fields list fields of a Adabas file
func Fields(clientInstance *client.AdabasAdmin, dbid int, fnr int, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewGetFieldDefinitionTableParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
	params.Dbid = float64(dbid)
	params.File = float64(fnr)
	resp, err := clientInstance.OnlineOffline.GetFieldDefinitionTable(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.GetFieldDefinitionTableBadRequest:
			response := err.(*online_offline.GetFieldDefinitionTableBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	p.Printf("\nDatabase %03d file %03d field definition table:\n", dbid, fnr)
	p.Println()
	p.Printf("Fields : %d\n", len(resp.Payload.FDT.Fields))
	p.Printf("Field Definition Table:\n")
	p.Println()
	p.Printf("   Level  I Name I Length I Format I   Options         I Flags   I Encoding\n")
	p.Printf("-------------------------------------------------------------------------------\n")
	for _, f := range resp.Payload.FDT.Fields {
		printFields(p, f)
	}
	p.Println()
	p.Println("Descriptors")
	p.Println("-------------------------------------------------------------------------------")
	p.Println("   Type   I Name I Length I Format I   Options         I Parent field(s)   Fmt")
	p.Println("-------------------------------------------------------------------------------")
	for _, f := range resp.Payload.FDT.Descriptors {
		printFields(p, f)
	}
	if len(resp.Payload.FDT.Referentials) > 0 {
		p.Println()
		p.Println("Referential Integrity")
		p.Println("-------------------------------------------------------------------------------")
		p.Println("	Type   I Name I Refer. I PrimaryI Foreign I Rules")
		p.Println("	       I      I file   I  field I  field  I")
		p.Println("-------------------------------------------------------------------------------")
		for _, f := range resp.Payload.FDT.Referentials {
			printFields(p, f)
		}
	}
	return nil
}

// AddFields add Adabas fields
func AddFields(clientInstance *client.AdabasAdmin, dbid int, fnr int, fdt string, auth runtime.ClientAuthInfoWriter) error {
	params := online_offline.NewModifyFieldDefinitionTableParams()
	params.Dbid = float64(dbid)
	params.File = float64(fnr)
	params.Addfields = fdt
	fmt.Println("Add fields", fdt)
	resp, err := clientInstance.OnlineOffline.ModifyFieldDefinitionTable(params, auth)
	if err != nil {
		switch err.(type) {
		case *online_offline.ModifyFieldDefinitionTableBadRequest:
			response := err.(*online_offline.ModifyFieldDefinitionTableBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Println("Status: ", resp.Payload.Status.Message)
	return nil

}

func printFields(p *message.Printer, f *models.Field) {
	space := bytes.Buffer{}
	for i := 0; i < int(f.Level); i++ {
		space.WriteString(" ")
	}
	after := bytes.Buffer{}
	for i := int(f.Level); i < 8; i++ {
		after.WriteString(" ")
	}
	switch f.Type {
	case "PHONETIC":
		p.Printf(" PHONETIC I   %3s I  %5s   \n",
			f.Name, f.Flags)
	case "FIELD":
		p.Printf("%s%d%s I %3s  I %6d  I %3s   I %s\n", space.String(), f.Level, after.String(),
			f.Name, f.Length, f.Format, f.Flags)
	case "COLLATION":
		p.Printf(" COLL     I    %-17s \n", f.Flags)
	case "SUB":
		fsf := (f.SubFields)[0]
		sf := *fsf
		p.Printf(" SUB      I   %3s I %5d I %5s  I %-17s I %3s(%5d,%5d)\n",
			f.Name, f.Length, f.Format, f.Flags, sf.SubName, sf.From, sf.To)
	case "SUPER":
		for i, s := range f.SubFields {
			if i == 0 {
				p.Printf(" SUPER    I   %3s I %5d I %5s  I %-17s I %3s(%5d,%5d)\n",
					f.Name, f.Length, f.Format, f.Flags, s.SubName, s.From, s.To)
			} else {
				p.Printf("          I       I       I        I %-17s I %3s(%5d,%5d)\n", " ", s.SubName, s.From, s.To)
			}
		}
	default:
		p.Printf(" PRIMARY       I  %3s I %6d I %3s    I %d\n",
			f.Name, f.Length, f.Format, f.Length)

	}

}
