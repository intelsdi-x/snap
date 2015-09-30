package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest/rbody"
	"github.com/intelsdi-x/pulse/mgmt/tribe"
	"github.com/intelsdi-x/pulse/scheduler"
)

var lock sync.Mutex = sync.Mutex{}

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
	req, err := http.NewRequest("POST", uri, b)
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
	lock.Lock()
	numOfNodes := 5
	aName := "agreement1"
	mgtPorts := startTribes(numOfNodes)
	lock.Unlock()
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
						resp := uploadPlugin(DUMMY_PLUGIN_PATH1, mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 201)
						So(resp.Meta.Type, ShouldEqual, rbody.PluginsLoadedType)
						resp = uploadPlugin(DUMMY_PUBLISHER_PATH, mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 201)
						So(resp.Meta.Type, ShouldEqual, rbody.PluginsLoadedType)
						resp = getPluginList(mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.PluginList).LoadedPlugins), ShouldEqual, 2)
						resp = getAgreement(mgtPorts[0], aName)
						So(resp.Meta.Code, ShouldEqual, 200)
						So(len(resp.Body.(*rbody.TribeGetAgreement).Agreement.PluginAgreement.Plugins), ShouldEqual, 2)
						resp = createTask("1.json", "task1", "1s", true, mgtPorts[0])
						So(resp.Meta.Code, ShouldEqual, 201)
						So(resp.Meta.Type, ShouldEqual, rbody.AddScheduledTaskType)
						So(resp.Body, ShouldHaveSameTypeAs, new(rbody.AddScheduledTask))

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

							Convey("The task has been shared and loaded across the cluster", func(c C) {
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
												resp := getTasks(port)
												if resp.Meta.Code == 200 {
													if len(resp.Body.(*rbody.ScheduledTaskListReturned).ScheduledTasks) == 1 {
														log.Debugf("node %v has %d tasks", port, len(resp.Body.(*rbody.ScheduledTaskListReturned).ScheduledTasks))
														return
													}
													log.Debugf("node %v has %d tasks", port, len(resp.Body.(*rbody.ScheduledTaskListReturned).ScheduledTasks))
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
							})
						})
					})
				})
			})
		})
	})
}

func TestTribePluginAgreements(t *testing.T) {
	lock.Lock()
	var (
		lpName, lpType string
		lpVersion      int
	)
	numOfNodes := 5
	aName := "agreement1"
	mgtPorts := startTribes(numOfNodes)
	lock.Unlock()
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
						resp := uploadPlugin(DUMMY_PLUGIN_PATH1, mgtPorts[0])
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

func startTribes(count int) []int {

	seed := ""
	var wg sync.WaitGroup
	var mgtPorts []int
	for i := 0; i < count; i++ {
		mgtPort := getAvailablePort()
		mgtPorts = append(mgtPorts, mgtPort)
		tribePort := getAvailablePort()
		conf := tribe.DefaultConfig(fmt.Sprintf("member-%v", mgtPort), "127.0.0.1", tribePort, seed, mgtPort)
		conf.MemberlistConfig.PushPullInterval = 5 * time.Second
		if seed == "" {
			seed = fmt.Sprintf("%s:%d", "127.0.0.1", tribePort)
		}
		t, err := tribe.New(conf)
		if err != nil {
			panic(err)
		}

		c := control.New()
		c.RegisterEventHandler("tribe", t)
		c.Start()
		s := scheduler.New()
		s.SetMetricManager(c)
		s.RegisterEventHandler("tribe", t)
		s.Start()
		t.SetPluginCatalog(c)
		t.SetTaskManager(s)
		t.Start()
		r := New()
		r.BindMetricManager(c)
		r.BindTaskManager(s)
		r.BindTribeManager(t)
		r.Start(":" + strconv.Itoa(mgtPort))
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
						restLogger.Infof("num of members %v", len(resp.Body.(*rbody.TribeMemberList).Members))
						return
					}
				}
			}
		}(mgtPort)
	}
	wg.Wait()
	return mgtPorts
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
