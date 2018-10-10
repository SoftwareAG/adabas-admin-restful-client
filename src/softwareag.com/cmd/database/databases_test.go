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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckLogging(t *testing.T) {
	log := checkLogging("1")
	assert.Equal(t, "CB", log)
	log = checkLogging("248")
	assert.Equal(t, "IO,RB,SB,VB,OFF", log)
}

func TestCheckUserExits(t *testing.T) {
	userExit := checkUserexits("2")
	assert.Equal(t, "2", userExit)
	userExit = checkUserexits("3")
	assert.Equal(t, "1,2", userExit)
	userExit = checkUserexits(strconv.Itoa(aifSpaUserexit11))
	assert.Equal(t, "11", userExit)
	userExit = checkUserexits(strconv.Itoa(aifSpaUserexit11 | aifSpaUserexit14))
	assert.Equal(t, "11,14", userExit)
}

func TestInputList(t *testing.T) {

	var il InputList
	il.Set("AA")
	il.Set("BB")
	assert.Equal(t, "AA,BB", il.String())
}
