// +build medium

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

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/gomit"
	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/tribe_event"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/intelsdi-x/snap/mgmt/tribe"
	"github.com/intelsdi-x/snap/scheduler"
)

var (
	tribeLogger = restLogger.WithFields(log.Fields{
		"_module": "rest-tribe",
	})
)

func getMembers(port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/members", port))
	if err != nil {
		restLogger.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getMember(port int, name string) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/member/%s", port, name))
	if err != nil {
		restLogger.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getAgreements(port int) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements", port))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func getAgreement(port int, name string) *rbody.APIResponse {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements/%s", port, name))
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func deleteAgreement(port int, name string) *rbody.APIResponse {
	client := &http.Client{}
	uri := fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements/%s", port, name)
	req, err := http.NewRequest("DELETE", uri, nil)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func joinAgreement(port int, memberName, agreementName string) *rbody.APIResponse {
	ja, err := json.Marshal(struct {
		MemberName string `json:"member_name"`
	}{
		MemberName: memberName,
	})
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewReader(ja)
	client := &http.Client{}
	uri := fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements/%s/join", port, agreementName)
	req, err := http.NewRequest("PUT", uri, b)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func leaveAgreement(port int, memberName, agreementName string) *rbody.APIResponse {
	ja, err := json.Marshal(struct {
		MemberName string `json:"member_name"`
	}{
		MemberName: memberName,
	})
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewReader(ja)
	client := &http.Client{}
	uri := fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements/%s/leave", port, agreementName)
	req, err := http.NewRequest("DELETE", uri, b)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func addAgreement(port int, name string) *rbody.APIResponse {
	a, err := json.Marshal(struct {
		Name string
	}{Name: name})
	if err != nil {
		log.Fatal(err)
	}
	b := bytes.NewReader(a)
	client := &http.Client{}
	uri := fmt.Sprintf("http://127.0.0.1:%d/v1/tribe/agreements", port)
	req, err := http.NewRequest("POST", uri, b)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return getAPIResponse(resp)
}

func TestTribeTaskAgreements(t *testing.T) {
	log.SetLevel(log.WarnLevel)
	numOfNodes := 5
	aName := "agreement99"
	mgtPorts, tribePort, lpe := startTribes(numOfNodes, "")
	Convey("A cluster is started", t, func() {
		Convey("Members are retrieved", func() {
			for _, i := range mgtPorts {
				m := getMembers(i)
				So(m.Body, ShouldHaveSameTypeAs, new(rbody.TribeMemberList))
				So(len(m.Body.(*rbody.TribeMemberList).Members), ShouldEqual, numOfNodes)
			}
		})
		Convey("An agreement is added", func() {
			a := addAgreement(mgtPorts[0], aName)
			So(a.Body, ShouldHaveSameTypeAs, new(rbody.TribeAddAgreement))
			Convey("All members join the agreement", func() {
				for _, i := range mgtPorts {
					j := joinAgreement(mgtPorts[0], fmt.Sprintf("member-%d", i), aName)
					So(j.Meta.Code, ShouldEqual, 200)
					So(j.Body, ShouldHaveSameTypeAs, new(rbody.TribeJoinAgreement))
				}
				Convey("All members have joined the agreement", func(c C) {
					var wg sync.WaitGroup
					timedOut := false
					for _, i := range mgtPorts {
						timer := time.After(15 * time.Second)
						wg.Add(1)
						go func(port int, name string) {
							defer wg.Done()
							for {
								select {
								case <-timer:
									timedOut = true
									return
								default:
									resp := getMember(port, name)
									if resp.Meta.Code == 200 {
										c.So(resp.Body.(*rbody.TribeMemberShow), ShouldHaveSameTypeAs, new(rbody.TribeMemberShow))
										if resp.Body.(*rbody.TribeMemberShow).PluginAgreement == aName {
											return
										}
									}
									time.Sleep(200 * time.Millisecond)
								}
							}
						}(i, fmt.Sprintf("member-%d", i))
					}
					wg.Wait()
					So(timedOut, ShouldEqual, false)

					Convey("Plugins and a task are uploaded", func() {
						resp := uploadPlugin(MOCK_PLUGIN_PATH2, mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 201)
						So(resp.Meta.Type, ShouldEqual, rbody.PluginsLoadedType)
						resp = getPluginList(mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.PluginList).LoadedPlugins), ShouldEqual, 1)
						pluginToUnload := resp.Body.(*rbody.PluginList).LoadedPlugins[0]
						<-lpe.pluginAddEvent
						resp = getAgreement(mgtPorts[0], aName)
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.TribeGetAgreement).Agreement.PluginAgreement.Plugins), ShouldEqual, 1)

						Convey("The cluster agrees on plugins", func(c C) {
							var wg sync.WaitGroup
							timedOut := false
							for _, i := range mgtPorts {
								timer := time.After(15 * time.Second)
								wg.Add(1)
								go func(port int) {
									defer wg.Done()
									for {
										select {
										case <-timer:
											timedOut = true
											return
										default:
											resp := getPluginList(port)
											if resp.Meta.Code == 200 {
												c.So(resp.Body.(*rbody.PluginList), ShouldHaveSameTypeAs, new(rbody.PluginList))
												if len(resp.Body.(*rbody.PluginList).LoadedPlugins) == 1 {
													return
												}
											}
											time.Sleep(200 * time.Millisecond)
										}
									}
								}(i)
							}
							wg.Wait()
							So(timedOut, ShouldEqual, false)

							resp = createTask("3.json", "task1", "1s", true, mgtPorts[0])
							So(resp.Meta.Code, ShouldEqual, 201)
							So(resp.Meta.Type, ShouldEqual, rbody.AddScheduledTaskType)
							So(resp.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))
							taskID := resp.Body.(*rbody.AddScheduledTask).ID

							Convey("The cluster agrees on tasks", func(c C) {
								var wg sync.WaitGroup
								timedOut := false
								for _, i := range mgtPorts {
									timer := time.After(15 * time.Second)
									wg.Add(1)
									go func(port int, name string) {
										defer wg.Done()
										for {
											select {
											case <-timer:
												timedOut = true
												return
											default:
												resp := getAgreement(port, name)
												if resp.Meta.Code == 200 {
													c.So(resp.Body.(*rbody.TribeGetAgreement), ShouldHaveSameTypeAs, new(rbody.TribeGetAgreement))
													if len(resp.Body.(*rbody.TribeGetAgreement).Agreement.TaskAgreement.Tasks) == 1 {
														return
													}
												}
												time.Sleep(200 * time.Millisecond)
											}
										}
									}(i, aName)
								}
								wg.Wait()
								So(timedOut, ShouldEqual, false)
								Convey("The task is started", func() {
									resp := startTask(taskID, mgtPorts[0])
									So(resp.Meta.Code, ShouldEqual, 200)
									So(resp.Meta.Type, ShouldEqual, rbody.ScheduledTaskStartedType)
									Convey("The task is started on all members of the tribe", func(c C) {
										var wg sync.WaitGroup
										timedOut := false
										for i := 0; i < numOfNodes; i++ {
											timer := time.After(15 * time.Second)
											wg.Add(1)
											go func(port int) {
												defer wg.Done()
												for {
													select {
													case <-timer:
														timedOut = true
														return
													default:
														resp := getTask(taskID, port)
														if resp.Meta.Code == 200 {
															if resp.Body.(*rbody.ScheduledTaskReturned).State == core.TaskSpinning.String() || resp.Body.(*rbody.ScheduledTaskReturned).State == core.TaskFiring.String() {
																return
															}
															log.Debugf("port %v has task in state %v", port, resp.Body.(*rbody.ScheduledTaskReturned).State)
														} else {
															log.Debugf("node %v error getting task", port)
														}
														time.Sleep(400 * time.Millisecond)
													}
												}
											}(mgtPorts[i])
										}
										wg.Wait()
										So(timedOut, ShouldEqual, false)
										Convey("A new node joins the agreement", func() {
											mgtPort, _, _ := startTribes(1, fmt.Sprintf("127.0.0.1:%d", tribePort))
											j := joinAgreement(mgtPort[0], fmt.Sprintf("member-%d", mgtPort[0]), aName)
											mgtPorts = append(mgtPorts, mgtPort[0])
											So(j.Meta.Code, ShouldEqual, 200)
											So(j.Body, ShouldHaveSameTypeAs, new(rbody.TribeJoinAgreement))
											var wg sync.WaitGroup
											timedOut := false
											for _, i := range mgtPort {
												timer := time.After(15 * time.Second)
												wg.Add(1)
												go func(port int, name string) {
													defer wg.Done()
													for {
														select {
														case <-timer:
															timedOut = true
															return
														default:
															resp := getTask(taskID, port)
															if resp.Meta.Code == 200 {
																if resp.Body.(*rbody.ScheduledTaskReturned).State == core.TaskSpinning.String() || resp.Body.(*rbody.ScheduledTaskReturned).State == core.TaskFiring.String() {
																	return
																}
																log.Debugf("port %v has task in state %v", port, resp.Body.(*rbody.ScheduledTaskReturned).State)
															} else {
																log.Debugf("node %v error getting task", port)
															}
															time.Sleep(400 * time.Millisecond)
														}
													}
												}(i, fmt.Sprintf("member-%d", i))
											}
											wg.Wait()
											time.Sleep(1 * time.Second)
											So(timedOut, ShouldEqual, false)
											Convey("The task is stopped", func() {
												resp := stopTask(taskID, mgtPorts[0])
												So(resp.Meta.Code, ShouldEqual, 200)
												So(resp.Meta.Type, ShouldEqual, rbody.ScheduledTaskStoppedType)
												So(resp.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskStopped))
												var wg sync.WaitGroup
												timedOut := false
												for i := 0; i < numOfNodes; i++ {
													timer := time.After(15 * time.Second)
													wg.Add(1)
													go func(port int) {
														defer wg.Done()
														for {
															select {
															case <-timer:
																timedOut = true
																return
															default:
																resp := getTask(taskID, port)
																if resp.Meta.Code == 200 {
																	if resp.Body.(*rbody.ScheduledTaskReturned).State == core.TaskStopped.String() {
																		return
																	}
																}
																time.Sleep(400 * time.Millisecond)
															}
														}
													}(mgtPorts[i])
												}
												wg.Wait()
												So(timedOut, ShouldEqual, false)
												Convey("The task is removed", func() {
													for _, port := range mgtPorts {
														resp := getTask(taskID, port)
														So(resp.Meta.Code, ShouldEqual, 200)
														So(resp.Body.(*rbody.ScheduledTaskReturned).State, ShouldResemble, core.TaskStopped.String())
													}
													resp := removeTask(taskID, mgtPorts[0])
													So(resp.Meta.Code, ShouldEqual, 200)
													So(resp.Meta.Type, ShouldEqual, rbody.ScheduledTaskRemovedType)
													So(resp.Body, ShouldHaveSameTypeAs, new(rbody.ScheduledTaskRemoved))
													var wg sync.WaitGroup
													timedOut := false
													for i := 0; i < numOfNodes; i++ {
														timer := time.After(15 * time.Second)
														wg.Add(1)
														go func(port int) {
															defer wg.Done()
															for {
																select {
																case <-timer:
																	timedOut = true
																	return
																default:
																	resp := getTask(taskID, port)
																	if resp.Meta.Code == 404 {
																		return
																	}
																	time.Sleep(400 * time.Millisecond)
																}
															}
														}(mgtPorts[i])
													}
													wg.Wait()
													So(timedOut, ShouldEqual, false)
													Convey("The plugins are unloaded", func(c C) {
														resp := unloadPlugin(mgtPorts[0], pluginToUnload.Type, pluginToUnload.Name, pluginToUnload.Version)
														So(resp.Meta.Code, ShouldEqual, 200)
														So(resp.Meta.Type, ShouldEqual, rbody.PluginUnloadedType)
														So(resp.Body, ShouldHaveSameTypeAs, new(rbody.PluginUnloaded))
														var wg sync.WaitGroup
														timedOut := false
														for i := 0; i < numOfNodes; i++ {
															timer := time.After(15 * time.Second)
															wg.Add(1)
															go func(port int) {
																defer wg.Done()
																for {
																	select {
																	case <-timer:
																		timedOut = true
																		return
																	default:
																		resp = getPluginList(port)
																		c.So(resp.Meta.Code, ShouldEqual, 200)
																		if len(resp.Body.(*rbody.PluginList).LoadedPlugins) == 0 {
																			return
																		}
																		time.Sleep(400 * time.Millisecond)
																	}
																}
															}(mgtPorts[i])
														}
														wg.Wait()
														So(timedOut, ShouldEqual, false)
													})
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})
}

func TestTribePluginAgreements(t *testing.T) {
	var (
		lpName, lpType string
		lpVersion      int
	)
	numOfNodes := 5
	aName := "agreement1"
	mgtPorts, _, _ := startTribes(numOfNodes, "")
	Convey("A cluster is started", t, func() {
		Convey("Members are retrieved", func() {
			for _, i := range mgtPorts {
				m := getMembers(i)
				So(m.Body, ShouldHaveSameTypeAs, new(rbody.TribeMemberList))
				So(len(m.Body.(*rbody.TribeMemberList).Members), ShouldEqual, numOfNodes)
			}
		})
		Convey("An agreement is added", func() {
			a := addAgreement(mgtPorts[0], aName)
			So(a.Body, ShouldHaveSameTypeAs, new(rbody.TribeAddAgreement))
			Convey("All members join the agreement", func() {
				for _, i := range mgtPorts {
					j := joinAgreement(mgtPorts[0], fmt.Sprintf("member-%d", i), aName)
					So(j.Meta.Code, ShouldEqual, 200)
					So(j.Body, ShouldHaveSameTypeAs, new(rbody.TribeJoinAgreement))
				}
				Convey("All members have joined the agreement", func(c C) {
					var wg sync.WaitGroup
					timedOut := false
					for _, i := range mgtPorts {
						timer := time.After(15 * time.Second)
						wg.Add(1)
						go func(port int, name string) {
							defer wg.Done()
							for {
								select {
								case <-timer:
									timedOut = true
									return
								default:
									resp := getMember(port, name)
									if resp.Meta.Code == 200 {
										c.So(resp.Body.(*rbody.TribeMemberShow), ShouldHaveSameTypeAs, new(rbody.TribeMemberShow))
										if resp.Body.(*rbody.TribeMemberShow).PluginAgreement == aName {
											return
										}
									}
									time.Sleep(200 * time.Millisecond)
								}
							}
						}(i, fmt.Sprintf("member-%d", i))
					}
					wg.Wait()
					So(timedOut, ShouldEqual, false)

					Convey("A plugin is uploaded", func() {
						resp := uploadPlugin(MOCK_PLUGIN_PATH2, mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 201)
						So(resp.Meta.Type, ShouldEqual, rbody.PluginsLoadedType)
						lpName = resp.Body.(*rbody.PluginsLoaded).LoadedPlugins[0].Name
						lpVersion = resp.Body.(*rbody.PluginsLoaded).LoadedPlugins[0].Version
						lpType = resp.Body.(*rbody.PluginsLoaded).LoadedPlugins[0].Type
						resp = getPluginList(mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.PluginList).LoadedPlugins), ShouldEqual, 1)
						resp = getAgreement(mgtPorts[0], aName)
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.TribeGetAgreement).Agreement.PluginAgreement.Plugins), ShouldEqual, 1)

						Convey("The cluster agrees on plugins", func(c C) {
							var wg sync.WaitGroup
							timedOut := false
							for _, i := range mgtPorts {
								timer := time.After(15 * time.Second)
								wg.Add(1)
								go func(port int, name string) {
									defer wg.Done()
									for {
										select {
										case <-timer:
											timedOut = true
											return
										default:
											resp := getAgreement(port, name)
											if resp.Meta.Code == 200 {
												c.So(resp.Body.(*rbody.TribeGetAgreement), ShouldHaveSameTypeAs, new(rbody.TribeGetAgreement))
												if len(resp.Body.(*rbody.TribeGetAgreement).Agreement.PluginAgreement.Plugins) == 1 {
													return
												}
											}
											time.Sleep(200 * time.Millisecond)
										}
									}
								}(i, aName)
							}
							wg.Wait()
							So(timedOut, ShouldEqual, false)

							Convey("The plugins have been shared and loaded across the cluster", func(c C) {
								var wg sync.WaitGroup
								timedOut := false
								for _, i := range mgtPorts {
									timer := time.After(15 * time.Second)
									wg.Add(1)
									go func(port int) {
										defer wg.Done()
										for {
											select {
											case <-timer:
												timedOut = true
												return
											default:
												resp := getPluginList(port)
												if resp.Meta.Code == 200 {
													c.So(resp.Body.(*rbody.PluginList), ShouldHaveSameTypeAs, new(rbody.PluginList))
													if len(resp.Body.(*rbody.PluginList).LoadedPlugins) == 1 {
														return
													}
												}
												time.Sleep(200 * time.Millisecond)
											}
										}
									}(i)
								}
								wg.Wait()
								So(timedOut, ShouldEqual, false)

								Convey("A plugin is unloaded", func() {
									resp := unloadPlugin(mgtPorts[0], lpType, lpName, lpVersion)
									So(resp.Meta.Code, ShouldEqual, 200)
									So(resp.Meta.Type, ShouldEqual, rbody.PluginUnloadedType)
									resp = getPluginList(mgtPorts[0])
									So(resp.Meta.Code, ShouldEqual, 200)
									So(len(resp.Body.(*rbody.PluginList).LoadedPlugins), ShouldEqual, 0)

									Convey("The cluster unloads the plugin", func() {
										var wg sync.WaitGroup
										timedOut := false
										for i := 0; i < numOfNodes; i++ {
											timer := time.After(15 * time.Second)
											wg.Add(1)
											go func(port int) {
												defer wg.Done()
												for {
													select {
													case <-timer:
														timedOut = true
														return
													default:
														resp := getPluginList(port)
														if resp.Meta.Code == 200 {
															c.So(resp.Body.(*rbody.PluginList), ShouldHaveSameTypeAs, new(rbody.PluginList))
															if len(resp.Body.(*rbody.PluginList).LoadedPlugins) == 0 {
																return
															}
															tribeLogger.Debugf("member %v has %v plugins", port, len(resp.Body.(*rbody.PluginList).LoadedPlugins))
														}
														time.Sleep(200 * time.Millisecond)
													}
												}
											}(mgtPorts[i])
										}
										wg.Wait()
										So(timedOut, ShouldEqual, false)

										Convey("A node leaves the agreement", func() {
											leaveAgreement(mgtPorts[0], fmt.Sprintf("member-%d", mgtPorts[0]), aName)
											var wg sync.WaitGroup
											timedOut := false
											for _, i := range mgtPorts {
												timer := time.After(15 * time.Second)
												wg.Add(1)
												go func(port int, name string) {
													defer wg.Done()
													for {
														select {
														case <-timer:
															timedOut = true
															return
														default:
															resp := getAgreement(port, aName)
															if resp.Meta.Code == 200 {
																c.So(resp.Body.(*rbody.TribeGetAgreement), ShouldHaveSameTypeAs, new(rbody.TribeGetAgreement))
																if len(resp.Body.(*rbody.TribeGetAgreement).Agreement.Members) == numOfNodes-1 {
																	return
																}
															}
															time.Sleep(200 * time.Millisecond)
														}
													}
												}(i, fmt.Sprintf("member-%d", i))
											}
											wg.Wait()
											So(timedOut, ShouldEqual, false)

											Convey("The agreement is deleted", func() {
												d := deleteAgreement(mgtPorts[0], aName)
												So(d.Body, ShouldHaveSameTypeAs, new(rbody.TribeDeleteAgreement))
												So(d.Meta.Code, ShouldEqual, 200)

												Convey("All members delete the agreement", func(c C) {
													var wg sync.WaitGroup
													timedOut := false
													for _, i := range mgtPorts {
														timer := time.After(15 * time.Second)
														wg.Add(1)
														go func(port int) {
															defer wg.Done()
															for {
																select {
																case <-timer:
																	timedOut = true
																	return
																default:
																	resp := getAgreements(port)
																	if resp.Meta.Code == 200 {
																		c.So(resp.Body.(*rbody.TribeListAgreement), ShouldHaveSameTypeAs, new(rbody.TribeListAgreement))
																		if len(resp.Body.(*rbody.TribeListAgreement).Agreements) == 0 {
																			return
																		}
																	}
																	time.Sleep(200 * time.Millisecond)
																}
															}
														}(i)
													}
													wg.Wait()
													So(timedOut, ShouldEqual, false)
												})
											})
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})
}

type listenToSeedEvents struct {
	pluginAddEvent chan struct{}
}

func newListenToSeedEvents() *listenToSeedEvents {
	return &listenToSeedEvents{
		pluginAddEvent: make(chan struct{}),
	}
}

func (l *listenToSeedEvents) HandleGomitEvent(e gomit.Event) {
	switch e.Body.(type) {
	case *tribe_event.AddPluginEvent:
		l.pluginAddEvent <- struct{}{}
	}
}

// returns an array of the mgtports and the tribe port for the last node
func startTribes(count int, seed string) ([]int, int, *listenToSeedEvents) {
	var wg sync.WaitGroup
	var tribePort int
	var mgtPorts []int
	lpe := newListenToSeedEvents()
	for i := 0; i < count; i++ {
		mgtPort := getAvailablePort()
		mgtPorts = append(mgtPorts, mgtPort)
		tribePort = getAvailablePort()
		conf := tribe.GetDefaultConfig()
		conf.Name = fmt.Sprintf("member-%v", mgtPort)
		conf.BindAddr = "127.0.0.1"
		conf.BindPort = tribePort
		conf.Seed = seed
		conf.RestAPIPort = mgtPort
		//conf.MemberlistConfig.PushPullInterval = 5 * time.Second
		conf.MemberlistConfig.RetransmitMult = conf.MemberlistConfig.RetransmitMult * 2

		t, err := tribe.New(conf)
		if err != nil {
			panic(err)
		}

		if seed == "" {
			seed = fmt.Sprintf("%s:%d", "127.0.0.1", tribePort)
			t.EventManager.RegisterHandler("tribe.tests", lpe)
		}

		cfg := control.GetDefaultConfig()
		// get an available port to avoid conflicts (we aren't testing remote workflows here)
		cfg.ListenPort = getAvailablePort()
		c := control.New(cfg)
		c.RegisterEventHandler("tribe", t)
		c.Start()
		s := scheduler.New(scheduler.GetDefaultConfig())
		s.SetMetricManager(c)
		s.RegisterEventHandler("tribe", t)
		s.Start()
		t.SetPluginCatalog(c)
		t.SetTaskManager(s)
		t.Start()
		r, _ := New(GetDefaultConfig())
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
					if resp.Meta.Code == 200 && len(resp.Body.(*rbody.TribeMemberList).Members) >= count {
						restLogger.Infof("num of members %v", len(resp.Body.(*rbody.TribeMemberList).Members))
						return
					}
				}
			}
		}(mgtPort)
	}
	wg.Wait()
	return mgtPorts, tribePort, lpe
}

var nextPort uint64 = 51234

func getAvailablePort() int {
	atomic.AddUint64(&nextPort, 1)
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", nextPort))
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()
	if err != nil {
		return getAvailablePort()
	}

	return l.Addr().(*net.TCPAddr).Port
}
