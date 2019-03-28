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

package job

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-openapi/runtime"
	"softwareag.com/client"
	"softwareag.com/client/scheduler"
	"softwareag.com/models"
)

// List list the jobs
func List(clientInstance *client.AdabasAdmin, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewGetJobsParams()
	resp, err := clientInstance.Scheduler.GetJobs(params, auth)
	if err != nil {
		switch err.(type) {
		case *scheduler.GetJobsNotFound:
			response := err.(*scheduler.GetJobsNotFound)
			fmt.Println(response.Error())
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf("Name             User        Status     Description\n")
	for _, j := range resp.Payload.JobDefinition {
		fmt.Printf("\n%-15s  %-8s    %-8s   %s\n", j.Job.Name, j.Job.User, j.Status, j.Job.Description)
		fmt.Println("  Executions:")
		for _, e := range j.Executions {
			fmt.Printf("    Id=%8d   Started at %8s ended at %s\n", e.ID, e.Scheduled, e.Ended)
		}
	}
	fmt.Println()
	return nil
}

// Start the job
func Start(clientInstance *client.AdabasAdmin, param string, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewScheduleJobParams()
	params.JobName = param
	resp, err := clientInstance.Scheduler.ScheduleJob(params, auth)
	if err != nil {
		switch err.(type) {
		case *scheduler.ScheduleJobBadRequest:
			response := err.(*scheduler.ScheduleJobBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Printf("Status message    : %s\n", resp.Payload.Status.Message)
	fmt.Printf("Job Name          : %s\n", resp.Payload.Status.Name)
	fmt.Printf("Execution ID      : %d\n", resp.Payload.Status.ExecutionID)
	fmt.Println()
	return nil
}

// Delete the job
func Delete(clientInstance *client.AdabasAdmin, param string, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewDeleteJobParams()
	params.JobName = param
	resp, err := clientInstance.Scheduler.DeleteJob(params, auth)
	if err != nil {
		switch err.(type) {
		case *scheduler.DeleteJobBadRequest:
			response := err.(*scheduler.DeleteJobBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf("Status message    : %s\n", resp.Payload.Status.Message)
	fmt.Println()
	return nil
}

// DeleteExecution Delete the execution log of a job
func DeleteExecution(clientInstance *client.AdabasAdmin, param string, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewDeleteJobResultParams()
	p := strings.Split(param, ":")
	params.JobName = p[0]
	params.JobID = p[1]
	resp, err := clientInstance.Scheduler.DeleteJobResult(params, auth)
	if err != nil {
		switch err.(type) {
		case *scheduler.DeleteJobResultBadRequest:
			response := err.(*scheduler.DeleteJobResultBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println()
	fmt.Printf("Status message    : %s\n", resp.Payload.Status.Message)
	fmt.Println()
	return nil
}

// Create nre job using job definition file
func Create(clientInstance *client.AdabasAdmin, input string, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewPostJobParams()
	raw, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	job := &models.JobParameter{}
	if err := json.Unmarshal(raw, job); err != nil {
		return err
	}
	params.Job = job
	resp, err := clientInstance.Scheduler.PostJob(params, auth)
	if err != nil {
		switch err.(type) {
		case *scheduler.PostJobBadRequest:
			response := err.(*scheduler.PostJobBadRequest)
			fmt.Println(response.Payload.Error.Code, ":", response.Payload.Error.Message)
		default:
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println(resp.Payload.Status.Message)
	return nil
}

// Log output
func Log(clientInstance *client.AdabasAdmin, param string, auth runtime.ClientAuthInfoWriter) error {
	params := scheduler.NewGetJobResultParams()
	v := strings.Split(param, ":")
	params.JobName = v[0]
	if len(v) != 2 {
		fmt.Printf("Parameter should be of form: <job name>:<execution id>\n")
		return fmt.Errorf("Parameter should be of form: <job name>:<execution id>")
	}
	params.JobID = v[1]
	resp, err := clientInstance.Scheduler.GetJobResult(params, auth)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println()
	fmt.Printf("JOB name     : %s\n", resp.Payload.JobResult.Name)
	fmt.Printf("JOB id       : %.0f\n", resp.Payload.JobResult.ID)
	fmt.Printf("JOB started  : %s\n", resp.Payload.JobResult.Scheduled)
	fmt.Printf("JOB ended    : %s\n", resp.Payload.JobResult.Ended)
	fmt.Printf("Output started -------:\n %s\nOutput ended -------\n", resp.Payload.JobResult.Log)
	fmt.Println()
	return nil
}
