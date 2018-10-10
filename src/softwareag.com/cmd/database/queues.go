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
	"fmt"

	"github.com/go-openapi/runtime"
	"softwareag.com/client"
	"softwareag.com/client/online"
)

// UserQueue display all user queue entries
func UserQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) {
	params := online.NewGetDatabaseUserQueueParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseUserQueue(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseUserQueueBadRequest:
			response := err.(*online.GetDatabaseUserQueueBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return
	}
	userQueue := resp.Payload.UserQueue

	fmt.Println()
	fmt.Println(" User queue entries:")
	fmt.Println()
	fmt.Printf(" %3s %-10s %-8s %-8s %-28s %-8s %-8s %-8s\n", "Id", "Es ID", "Node Id", "Login Id", "Timestamp", "User", "Flags", "ETFlags")
	for _, u := range userQueue.UserQueueEntry {
		fmt.Printf(" %3d %10d %-8s %-8s %-8s %-8s %-8s %-8s\n", u.UqID, u.UID.ID, u.UID.Node, u.UID.Terminal,
			u.UID.Timestamp, u.User, u.Flags, u.EtFlags)
	}
}

// CommandQueue display all command queue entries
func CommandQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) {
	params := online.NewGetDatabaseCommandQueueParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseCommandQueue(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseCommandQueueBadRequest:
			response := err.(*online.GetDatabaseCommandQueueBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return
	}

	fmt.Println()
	fmt.Println(" Command queue entries:")
	fmt.Println()
	fmt.Printf(" %3s  %-8s  %-8s  %-10s  %-3s  %-8s  %-8s\n", "No", "Node Id", "Login Id", "ES Id", "Cmd", "File", "Status")
	for _, c := range resp.Payload.CommandQueue.Commands {
		fmt.Printf(" %3d  %-8s  %-8s  %-10d  %-3s  %-8d  %-s\n", c.CommID, c.User.Node, c.User.Terminal, c.User.ID, c.CommandCode, c.File, c.Flags)
	}
}

// HoldQueue display all hold queue entries
func HoldQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) {
	params := online.NewGetDatabaseHoldQueueParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseHoldQueue(params, auth)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println()
	fmt.Println(" Hold queue entries:")
	fmt.Println()
	fmt.Printf("   Id Node Id   Login Id     ES Id     User Id  File           ISN Locks  Flg\n")
	for _, c := range resp.Payload.HoldQueueEntry {
		fmt.Printf(" %3d  %-8s  %-8s     %3d  %3s  %d  %s %s\n", c.HqCommid, c.Hid.Node, c.Hid.Terminal, c.Hid.ID, c.User, c.File, c.Locks, c.Flags)
	}
}
