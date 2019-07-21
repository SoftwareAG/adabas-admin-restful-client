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
	"strconv"

	"github.com/go-openapi/runtime"
	"softwareag.com/client"
	"softwareag.com/client/online"
)

// UserQueue display all user queue entries
func UserQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseUserQueueParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
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
		return err
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
	return nil
}

// UserDetails retrieve user queue entry details
func UserDetails(clientInstance *client.AdabasAdmin, dbid int, param string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetUserQueueDetailParams()
	params.Dbid = float64(dbid)
	qid, qerr := strconv.Atoi(param)
	if qerr != nil {
		return qerr
	}
	params.Queueid = float64(qid)

	resp, err := clientInstance.Online.GetUserQueueDetail(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseCommandQueueBadRequest:
			response := err.(*online.GetUserQueueDetailBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	userDetails := resp.Payload.UserQueueDetail.DetailEntry[0]
	fmt.Printf(" Got user queue details of queue id %v:\n", userDetails.UqID)
	fmt.Printf("%20s : %s\n", "User", userDetails.User)
	fmt.Printf("%20s :\n", "Adabas ID")
	fmt.Printf("%21s : %d\n", " ID", userDetails.UID.ID)
	fmt.Printf("%21s : %s\n", " Node", userDetails.UID.Node)
	fmt.Printf("%21s : %s\n", " Terminal", userDetails.UID.Terminal)
	fmt.Printf("%21s : %s\n", " Timestamp", userDetails.UID.Timestamp)
	fmt.Printf("%20s : %s\n", "Flags", userDetails.Flags)
	fmt.Printf("%20s : %s\n", "ET Flags", userDetails.EtFlags)
	fmt.Printf("%20s : %s\n", "Start session", resp.Payload.StartSession)
	fmt.Printf("%20s : %s\n", "Start transaction", resp.Payload.StartTransaction)
	fmt.Printf("%20s : %s\n", "Last activity", resp.Payload.LastActivity)
	fmt.Printf("%20s : %d\n", "TT Limit", resp.Payload.TTLimit)
	fmt.Printf("%20s : %d\n", "TNA Limit", resp.Payload.TNALimit)
	fmt.Printf("%20s : %d\n", "ISN lists", resp.Payload.ISNLists)
	fmt.Printf("%20s : %d\n", "ISN in hold", resp.Payload.ISNHold)
	fmt.Printf("%20s :\n", "Files in use")
	for f := range resp.Payload.Files {
		if f > 0 {
			fmt.Printf("%20s : %d\n", " ", f)
		}
	}
	fmt.Printf("%20s : %d\n", "Command count:", resp.Payload.CommandCount)
	fmt.Printf("%20s : %d\n", "Transaction count:", resp.Payload.TransactionCount)
	fmt.Printf("%20s : %d\n", "User encoding:", resp.Payload.UserEncoding)
	fmt.Println()
	return nil
}

// DeleteUser stop user
func DeleteUser(clientInstance *client.AdabasAdmin, dbid int, param string, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewStopUserQueueEntryParams()
	params.Dbid = float64(dbid)
	qid, qerr := strconv.Atoi(param)
	if qerr != nil {
		return qerr
	}
	params.Queueid = float64(qid)

	resp, err := clientInstance.Online.StopUserQueueEntry(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseCommandQueueBadRequest:
			response := err.(*online.StopUserQueueEntryBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf(" Stop of user %v in user queue initiated\n", params.Queueid)
	fmt.Println()
	fmt.Println(resp)
	return nil
}

// CommandQueue display all command queue entries
func CommandQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseCommandQueueParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
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
		return err
	}

	fmt.Println()
	fmt.Println(" Command queue entries:")
	fmt.Println()
	fmt.Printf(" %3s  %-8s  %-8s  %-10s  %-3s  %-8s  %-8s\n", "No", "Node Id", "Login Id", "ES Id", "Cmd", "File", "Status")
	for _, c := range resp.Payload.CommandQueue.Commands {
		fmt.Printf(" %3d  %-8s  %-8s  %-10d  %-3s  %-8d  %-s\n", c.CommID, c.User.Node, c.User.Terminal, c.User.ID, c.CommandCode, c.File, c.Flags)
	}
	return nil
}

// HoldQueue display all hold queue entries
func HoldQueue(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseHoldQueueParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseHoldQueue(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseHoldQueueBadRequest:
			response := err.(*online.GetDatabaseHoldQueueBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Println()
	fmt.Println(" Hold queue entries:")
	fmt.Println()
	fmt.Printf("   Id Node Id   Login Id     ES Id     User Id  File           ISN Locks  Flg\n")
	for _, c := range resp.Payload.HoldQueue {
		fmt.Printf(" %3d  %-8s  %-8s     %3d  %3s  %-3d  %d %s %s\n", c.HqCommid, c.Hid[0].Node, c.Hid[0].Terminal, c.Hid[0].ID, c.User, c.File, c.Isn, c.Locks, c.Flags)
	}
	return nil
}
