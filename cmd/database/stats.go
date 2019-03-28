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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"softwareag.com/client"
	"softwareag.com/client/online"
)

// Highwater High water statistics
func Highwater(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseHighWaterParams()
	rfc3339 := true
	params.Rfc3339 = &rfc3339
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseHighWater(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseHighWaterBadRequest:
			response := err.(*online.GetDatabaseHighWaterBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	fmt.Println()
	fmt.Printf("Database %d, startup at %s\n", dbid, resp.Payload.HighWater.NucleusStartTime)
	fmt.Println("High Water Mark:")
	fmt.Println()
	p.Printf("%-18s  %10s   %10s   %10s   %02s  %s\n", "Area/Entry", "Size", "In Use", "High Water", "%", "Date/Time")
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "User Queue", resp.Payload.HighWater.UserQueueSize,
		resp.Payload.HighWater.UserQueueHighWaterMark.Inuse, resp.Payload.HighWater.UserQueueHighWaterMark.High, 0,
		resp.Payload.HighWater.UserQueueHighWaterMark.Time)
	p.Printf("%-18s  %10s   %10d   %10d   %02d  %s\n", "Command Queue", "-",
		resp.Payload.HighWater.CommandQueueHighWaterMark.Inuse, resp.Payload.HighWater.CommandQueueHighWaterMark.High, 0,
		resp.Payload.HighWater.CommandQueueHighWaterMark.Time)
	p.Printf("%-18s  %10s   %10d   %10d   %02d  %s\n", "Hold Queue", "-",
		resp.Payload.HighWater.HoldQueueHighWaterMark.Inuse, resp.Payload.HighWater.HoldQueueHighWaterMark.High, 0,
		resp.Payload.HighWater.HoldQueueHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Client Queue", resp.Payload.HighWater.ClientQueueSize,
		resp.Payload.HighWater.ClientQueueHighWaterMark.Inuse, resp.Payload.HighWater.ClientQueueHighWaterMark.High, 0,
		resp.Payload.HighWater.ClientQueueHighWaterMark.Time)
	p.Printf("%-18s  %10s   %10d   %10d   %02d  %s\n", "HQ User Limit", "-",
		resp.Payload.HighWater.HQUserLimitHighWaterMark.Inuse, resp.Payload.HighWater.HQUserLimitHighWaterMark.High, 0,
		resp.Payload.HighWater.UserQueueHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Threads", resp.Payload.HighWater.ThreadSize,
		resp.Payload.HighWater.ThreadsHighWaterMark.Inuse, resp.Payload.HighWater.ThreadsHighWaterMark.High, 0,
		resp.Payload.HighWater.ThreadsHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Workpool", resp.Payload.HighWater.WorkpoolSize,
		resp.Payload.HighWater.WorkpoolHighWaterMark.Inuse, resp.Payload.HighWater.WorkpoolHighWaterMark.High, 0,
		resp.Payload.HighWater.WorkpoolHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "  ISN Sort", resp.Payload.HighWater.SortAreaSize,
		resp.Payload.HighWater.IsnSortHighWaterMark.Inuse, resp.Payload.HighWater.IsnSortHighWaterMark.High, 0,
		resp.Payload.HighWater.IsnSortHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "  Complex Search", resp.Payload.HighWater.SortAreaSize,
		resp.Payload.HighWater.ComplexSearchHighWaterMark.Inuse, resp.Payload.HighWater.ComplexSearchHighWaterMark.High, 0,
		resp.Payload.HighWater.ComplexSearchHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Attached Buffer", resp.Payload.HighWater.AttachedBufferSize,
		resp.Payload.HighWater.AttachedBufferHighWaterMark.Inuse, resp.Payload.HighWater.AttachedBufferHighWaterMark.High, 0,
		resp.Payload.HighWater.AttachedBufferHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "ATBX (MB)", resp.Payload.HighWater.LABXSize,
		resp.Payload.HighWater.LABXHighWaterMark.Inuse, resp.Payload.HighWater.LABXHighWaterMark.High, 0,
		resp.Payload.HighWater.LABXHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Buffer Pool", resp.Payload.HighWater.BufferpoolSize,
		resp.Payload.HighWater.BufferpoolHighWaterMark.Inuse, resp.Payload.HighWater.BufferpoolHighWaterMark.High, 0,
		resp.Payload.HighWater.BufferpoolHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Protection Area", resp.Payload.HighWater.ProtectionAreaSize,
		resp.Payload.HighWater.WorkpoolHighWaterMark.Inuse, resp.Payload.HighWater.WorkpoolHighWaterMark.High, 0,
		resp.Payload.HighWater.WorkpoolHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "  Active Area", resp.Payload.HighWater.ProtectionAreaActiveSize,
		resp.Payload.HighWater.ProtectionAreaActiveHighWaterMark.Inuse, resp.Payload.HighWater.ProtectionAreaActiveHighWaterMark.High, 0,
		resp.Payload.HighWater.ProtectionAreaActiveHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Group Commit", resp.Payload.HighWater.GroupCommitSize,
		resp.Payload.HighWater.GroupCommitHighWaterMark.Inuse, resp.Payload.HighWater.GroupCommitHighWaterMark.High, 0,
		resp.Payload.HighWater.GroupCommitHighWaterMark.Time)
	p.Printf("%-18s  %10d   %10d   %10d   %02d  %s\n", "Transaction Commit", resp.Payload.HighWater.TransactionTimeSize,
		resp.Payload.HighWater.TransactionTimeHighWaterMark.Inuse, resp.Payload.HighWater.TransactionTimeHighWaterMark.High, 0,
		resp.Payload.HighWater.TransactionTimeHighWaterMark.Time)
	return nil
}

// CommandStats command statistics
func CommandStats(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseCommandStatsParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseCommandStats(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseCommandStatsBadRequest:
			response := err.(*online.GetDatabaseCommandStatsBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	fmt.Println()
	fmt.Println(" Adabas command statistics:")
	for i, c := range resp.Payload.CommandStats.Commands {
		if i%3 == 0 {
			fmt.Println()
		}
		fmt.Printf(" %3s  %8d\t\t", c.CommandName, c.CommandCount)
	}
	fmt.Println()
	return nil
}

// BufferpoolStats buffer pool statistics
func BufferpoolStats(clientInstance *client.AdabasAdmin, dbid int, auth runtime.ClientAuthInfoWriter) error {
	params := online.NewGetDatabaseBPStatsParams()
	params.Dbid = float64(dbid)
	resp, err := clientInstance.Online.GetDatabaseBPStats(params, auth)
	if err != nil {
		switch err.(type) {
		case *online.GetDatabaseBPStatsBadRequest:
			response := err.(*online.GetDatabaseBPStatsBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}

	p := message.NewPrinter(language.English)

	fmt.Println()
	fmt.Println(" Adabas buffer pool statistics:")
	fmt.Println()
	p.Printf(" Buffer Pool Size    :  %8d\n", resp.Payload.Statistics.Size)
	fmt.Println()
	fmt.Println(" Pool Allocation                        RABNs present")
	fmt.Println(" ---------------                        -------------")
	percent := resp.Payload.Statistics.AllocCurrent * 100 / resp.Payload.Statistics.Size
	p.Printf(" Current     (%3d%%) :  %12d     ASSO               : %12d\n", percent, resp.Payload.Statistics.AllocCurrent, resp.Payload.Statistics.RabnsAsso)
	percent = resp.Payload.Statistics.AllocHighwater * 100 / resp.Payload.Statistics.Size
	p.Printf(" Highwater   (%3d%%) :  %12d     DATA               : %12d\n", percent, resp.Payload.Statistics.AllocHighwater, resp.Payload.Statistics.RabnsData)
	percent = resp.Payload.Statistics.AllocInternal * 100 / resp.Payload.Statistics.Size
	p.Printf(" Internal    (%3d%%) :  %12d     WORK               : %12d\n", percent, resp.Payload.Statistics.AllocInternal, resp.Payload.Statistics.RabnsWork)
	percent = resp.Payload.Statistics.AllocWorkpool * 100 / resp.Payload.Statistics.Size
	p.Printf(" Workpool    (%3d%%) :  %12d     NUCTMP             : %12d\n", percent, resp.Payload.Statistics.AllocWorkpool, resp.Payload.Statistics.RabnsNucTmp)
	p.Printf("                                        NUCSRT             : %12d\n", resp.Payload.Statistics.RabnsNucSort)
	p.Printf("\n")
	p.Printf(" I/O Statistics                         Buffer Flushes\n")
	p.Printf(" --------------                         --------------\n")
	p.Printf(" Logical Reads      :  %12d     Total              : %12d\n", resp.Payload.Statistics.IOLogicalReads, resp.Payload.Statistics.FlushesTotal)
	p.Printf(" Physical Reads     :  %12d     To Free Space      : %12d\n", resp.Payload.Statistics.IOPhysicalsReads, resp.Payload.Statistics.FlushesFree)
	phitrate := float64(resp.Payload.Statistics.IOLogicalReads-resp.Payload.Statistics.IOPhysicalsReads) / float64(resp.Payload.Statistics.IOLogicalReads) * 100
	p.Printf(" Pool Hit Rate      :            %.1f%%  Temporary Blocks   : %12d\n", phitrate, 0)

	p.Printf("                                        Write Limit  ( 50%%): %12d\n", resp.Payload.Statistics.WriteLimit)
	p.Printf(" Physical Writes    :  %12d     Modified     (  0%%): %12d\n", resp.Payload.Statistics.IOPhysicalWrites, resp.Payload.Statistics.Modified)

	//fmt.Printf("                                        Limit Temp.B.( 50%%): %12d\n", 0)
	//fmt.Printf("                                        Modified T.B.(  0%%): %12d\n", 0)
	fmt.Println()
	return nil
}
