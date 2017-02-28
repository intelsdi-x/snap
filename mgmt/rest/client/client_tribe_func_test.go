// +build legacy

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/mgmt/rest"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/intelsdi-x/snap/mgmt/tribe"
	"github.com/intelsdi-x/snap/scheduler"
)

func getPort() int {
	// This attempts to use net.Listen to find an open port since
	// the tribe config has to know of one BEFORE the REST API starts...
	count := 0
	// This will loop 1000 times before panicking
	// If it finds a port it will return out of the function
	for count < 1000 {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err == nil {
			// Grab port from listener
			p := ln.Addr().(*net.TCPAddr).Port
			ln.Close()
			return p
		}
		count++
	}
	// We tried 1000 times and just give up
	panic("Could not get a port")
}

func readBody(r *http.Response) []byte {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()
	return b
}

func getAPIResponse(resp *http.Response) *rbody.APIResponse {
	r := new(rbody.APIResponse)
	rb := readBody(resp)
	err := json.Unmarshal(rb, r)
	if err != nil {
		log.Fatal(err)
	}
	r.JSONResponse = string(rb)
	return r
}

func getMembers(port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/members", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func TestSnapClientTribe(t *testing.T) {
	numOfTribes := 4
	ports := startTribes(numOfTribes)
	c, err := client.New(fmt.Sprintf("http://localhost:%d", ports[0]), "v1", true)

	Convey("REST API functional V1 - TRIBE", t, func() {
		So(err, ShouldBeNil)

		Convey("Get global membership", func() {
			resp := c.ListMembers()
			So(resp.Err, ShouldBeNil)
			So(resp.Members, ShouldNotBeNil)
			So(resp, ShouldHaveSameTypeAs, &client.ListMembersResult{})
			So(len(resp.Members), ShouldEqual, numOfTribes)
		})

		Convey("An agreement is added", func() {
			agreement := "agreement1"
			resp := c.AddAgreement(agreement)
			So(resp.Err, ShouldBeNil)
			So(resp, ShouldHaveSameTypeAs, &client.AddAgreementResult{})
			So(len(resp.Agreements), ShouldEqual, 1)
			resp2 := c.ListAgreements()
			So(resp2.Err, ShouldBeNil)
			So(resp2, ShouldHaveSameTypeAs, &client.ListAgreementResult{})
			So(len(resp2.Agreements), ShouldEqual, 1)
			Convey("A node joins the agreement", func() {
				resp := c.JoinAgreement(agreement, fmt.Sprintf("member-%d", ports[0]))
				So(resp.Err, ShouldBeNil)
				So(resp, ShouldHaveSameTypeAs, &client.JoinAgreementResult{})
				So(resp.Agreement, ShouldNotBeNil)
				So(len(resp.Agreement.Members), ShouldEqual, 1)
				Convey("The rest of the members join the agreement", func() {
					for i := 1; i < numOfTribes; i++ {
						resp := c.JoinAgreement(agreement, fmt.Sprintf("member-%d", ports[i]))
						So(resp.Err, ShouldBeNil)
						So(resp, ShouldHaveSameTypeAs, &client.JoinAgreementResult{})
						So(resp.Agreement, ShouldNotBeNil)
						So(len(resp.Agreement.Members), ShouldEqual, i+1)
					}
					Convey("A member is removed from the agreement", func() {
						resp := c.LeaveAgreement(agreement, fmt.Sprintf("member-%d", ports[0]))
						So(resp.Err, ShouldBeNil)
						So(resp, ShouldHaveSameTypeAs, &client.LeaveAgreementResult{})
						So(resp.Agreement, ShouldNotBeNil)
						So(len(resp.Agreement.Members), ShouldEqual, numOfTribes-1)
						Convey("A member is retrieved", func() {
							resp := c.GetMember(fmt.Sprintf("member-%d", ports[1]))
							So(resp.Err, ShouldBeNil)
							So(resp, ShouldHaveSameTypeAs, &client.GetMemberResult{})
							So(resp.Name, ShouldNotBeNil)
							So(resp.Name, ShouldResemble, fmt.Sprintf("member-%d", ports[1]))
							Convey("An Agreement is retrieved", func() {
								resp := c.GetAgreement(agreement)
								So(resp.Err, ShouldBeNil)
								So(resp, ShouldHaveSameTypeAs, &client.GetAgreementResult{})
								So(resp.Agreement.Name, ShouldNotBeNil)
								So(resp.Agreement.Name, ShouldResemble, agreement)
								So(len(resp.Agreement.Members), ShouldEqual, 3)
								Convey("An agreement is deleted", func() {
									resp := c.DeleteAgreement(agreement)
									So(resp.Err, ShouldBeNil)
									So(resp, ShouldHaveSameTypeAs, &client.DeleteAgreementResult{})
									So(len(resp.Agreements), ShouldEqual, 0)
								})
							})
						})
					})
				})
			})
		})
	})
}

func startTribes(count int) []int {
	seed := ""
	var wg sync.WaitGroup
	var mgtPorts []int
	for i := 0; i < count; i++ {
		mgtPort := getPort()
		mgtPorts = append(mgtPorts, mgtPort)
		tribePort := getPort()
		conf := tribe.GetDefaultConfig()
		conf.Name = fmt.Sprintf("member-%v", mgtPort)
		conf.BindAddr = "127.0.0.1"
		conf.BindPort = tribePort
		conf.Seed = seed
		conf.RestAPIPort = mgtPort
		conf.MemberlistConfig.PushPullInterval = 5 * time.Second
		conf.MemberlistConfig.RetransmitMult = conf.MemberlistConfig.RetransmitMult * 2
		if seed == "" {
			seed = fmt.Sprintf("%s:%d", "127.0.0.1", tribePort)
		}
		t, err := tribe.New(conf)
		if err != nil {
			panic(err)
		}

		c := control.New(control.GetDefaultConfig())
		c.RegisterEventHandler("tribe", t)
		c.Start()
		s := scheduler.New(scheduler.GetDefaultConfig())
		s.SetMetricManager(c)
		s.RegisterEventHandler("tribe", t)
		s.Start()
		t.SetPluginCatalog(c)
		t.SetTaskManager(s)
		t.Start()
		r, _ := rest.New(rest.GetDefaultConfig())
		r.BindMetricManager(c)
		r.BindTaskManager(s)
		r.BindTribeManager(t)
		r.SetAddress(fmt.Sprintf("127.0.0.1:%d", mgtPort))
		r.Start()
		wg.Add(1)
		timer := time.After(10 * time.Second)
		go func(port int) {
			defer wg.Done()
			for {
				select {
				case <-timer:
					panic("timed out")
				default:
					time.Sleep(100 * time.Millisecond)

					resp := getMembers(port)
					if resp.Meta.Code == 200 && len(resp.Body.(*rbody.TribeMemberList).Members) == count {
						log.Infof("num of members %v", len(resp.Body.(*rbody.TribeMemberList).Members))
						return
					}
				}
			}
		}(mgtPort)
	}
	wg.Wait()
	uris := make([]int, len(mgtPorts))
	for idx, port := range mgtPorts {
		uris[idx] = port
	}
	return uris
}
