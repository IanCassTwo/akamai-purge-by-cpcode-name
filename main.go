/*
 * Copyright 2018. Akamai Technologies, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"bytes"
	"log"
	"encoding/json"
        "github.com/akamai/AkamaiOPEN-edgegrid-golang/edgegrid"
        "github.com/akamai/AkamaiOPEN-edgegrid-golang/client-v1"

) 

type Cpcodes struct {
	Cpcodes []struct {
		CpcodeID         int    `json:"cpcodeId"`
		CpcodeName       string `json:"cpcodeName"`
		Purgeable        bool   `json:"purgeable"`
	} `json:"cpcodes"`
}

type PurgeRequest struct {
	Objects []int `json:"objects"`
}

type PurgeResponse struct {
	HTTPStatus       int    `json:"httpStatus"`
	EstimatedSeconds int    `json:"estimatedSeconds"`
	PurgeID          string `json:"purgeId"`
	SupportID        string `json:"supportId"`
	Detail           string `json:"detail"`
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: ", os.Args[0], " <cpcode name>", " [network]")
	}

	var network = ""
	if (len(os.Args) == 3 && os.Args[2] == "staging") {
		network = fmt.Sprintf("/%s", os.Args[2])
	}

	// Our regular API key. Needs read permissions for "CPcode and Reporting group (cprg)"
        config, err := edgegrid.Init("~/.edgerc", "default")
        if err != nil {
		log.Fatal(err)
        }

	// Unfortunately, you can't purge with a regular API key. You need a specific "purge" key
        ccuconfig, err := edgegrid.Init("~/.edgerc", "ccu")
        if err != nil {
		log.Fatal(err)
        }

	// Request a list of CPCodes
	req, err := client.NewRequest(config, "GET", "/cprg/v1/cpcodes", nil)
	resp, _ := client.Do(config, req)
        if err != nil {
		log.Fatal(err)
        }

        byt, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if (resp.StatusCode != 200) {
		fmt.Println(string(byt))
		os.Exit(1)
	}

        var cpcodes Cpcodes
        err = json.Unmarshal(byt, &cpcodes)

	// Iterate CPCodes to find the one that we need
	for _, item := range cpcodes.Cpcodes {
		if item.CpcodeName == os.Args[1] {

			// Code found - let's purge it
			var purgerequest PurgeRequest
			purgerequest.Objects = append(purgerequest.Objects, item.CpcodeID)
			p,_ := json.Marshal(purgerequest)

			req, _ = client.NewRequest(ccuconfig, "POST", fmt.Sprintf("/ccu/v3/invalidate/cpcode%s", network), bytes.NewBuffer(p))
			resp, err = client.Do(ccuconfig, req)
			if err != nil {
				log.Fatal(err)
			}

			byt, _ = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()

			if (resp.StatusCode != 201) {
				fmt.Println(string(byt))
				os.Exit(1)
			}

			var purgeresponse PurgeResponse
        		err = json.Unmarshal(byt, &purgeresponse)

			s,_ := json.MarshalIndent(purgeresponse, "", "\t")
			fmt.Println(string(s))
	
			os.Exit(0)
		}
	}

	// Only get here if we couldn't find out code
	fmt.Printf("Couldn't find cpcode for %s\n", os.Args[1])

}
