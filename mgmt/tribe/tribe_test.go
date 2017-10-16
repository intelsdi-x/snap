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

package tribe

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/serror"
	"github.com/intelsdi-x/snap/mgmt/tribe/agreement"
	"github.com/intelsdi-x/snap/pkg/schedule"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

type mockTaskManager struct{}

func (m *mockTaskManager) GetTask(id string) (core.Task, error) { return &mockTask{}, nil }
func (m *mockTaskManager) CreateTaskTribe(sch schedule.Schedule, wfMap *wmap.WorkflowMap, startOnCreate bool, opts ...core.TaskOption) (core.Task, core.TaskErrors) {
	return nil, nil
}
func (m *mockTaskManager) StopTaskTribe(id string) []serror.SnapError  { return nil }
func (m *mockTaskManager) StartTaskTribe(id string) []serror.SnapError { return nil }
func (m *mockTaskManager) RemoveTaskTribe(id string) error             { return nil }

type mockTask struct{}

func (t *mockTask) ID() string                                { return "" }
func (t *mockTask) State() core.TaskState                     { return core.TaskSpinning }
func (t *mockTask) HitCount() uint                            { return 0 }
func (t *mockTask) GetName() string                           { return "" }
func (t *mockTask) SetName(string)                            { return }
func (t *mockTask) SetID(string)                              { return }
func (t *mockTask) MissedCount() uint                         { return 0 }
func (t *mockTask) FailedCount() uint                         { return 0 }
func (t *mockTask) LastFailureMessage() string                { return "" }
func (t *mockTask) LastRunTime() *time.Time                   { return nil }
func (t *mockTask) CreationTime() *time.Time                  { return nil }
func (t *mockTask) DeadlineDuration() time.Duration           { return 0 }
func (t *mockTask) SetDeadlineDuration(time.Duration)         { return }
func (t *mockTask) SetTaskID(id string)                       { return }
func (t *mockTask) SetStopOnFailure(int)                      { return }
func (t *mockTask) GetStopOnFailure() int                     { return 0 }
func (t *mockTask) Option(...core.TaskOption) core.TaskOption { return core.TaskDeadlineDuration(0) }
func (t *mockTask) WMap() *wmap.WorkflowMap                   { return nil }
func (t *mockTask) Schedule() schedule.Schedule               { return nil }
func (t *mockTask) MaxFailures() int                          { return 10 }
func (t *mockTask) MaxMetricsBuffer() int64                   { return 0 }
func (t *mockTask) SetMaxMetricsBuffer(int64)                 {}
func (t *mockTask) MaxCollectDuration() time.Duration         { return time.Second }
func (t *mockTask) SetMaxCollectDuration(time.Duration)       {}

func getTestConfig() *Config {
	cfg := GetDefaultConfig()
	cfg.BindAddr = "127.0.0.1"
	cfg.BindPort = getAvailablePort()
	cfg.RestAPIPort = getAvailablePort()
	cfg.MemberlistConfig.PushPullInterval = 200 * time.Millisecond
	cfg.MemberlistConfig.GossipInterval = 300 * time.Second
	cfg.MemberlistConfig.GossipNodes = 0
	return cfg
}

func TestTribeFullStateSync(t *testing.T) {
	log.SetLevel(log.WarnLevel)
	tribes := []*tribe{}
	numOfTribes := 5
	agreement1 := "agreement1"
	plugin1 := agreement.Plugin{Name_: "plugin1", Version_: 1, Type_: core.ProcessorPluginType}
	task1 := agreement.Task{ID: uuid.New()}
	Convey("Tribe members are started", t, func() {
		conf := getTestConfig()
		conf.Name = "seed"
		seed, err := New(conf)
		So(seed, ShouldNotBeNil)
		So(err, ShouldBeNil)
		taskManager := &mockTaskManager{}
		seed.SetTaskManager(taskManager)
		tribes = append(tribes, seed)
		for i := 1; i < numOfTribes; i++ {
			conf := getTestConfig()
			conf.Name = fmt.Sprintf("member-%v", i)
			conf.Seed = fmt.Sprintf("%v:%v", "127.0.0.1", seed.memberlist.LocalNode().Port)
			tr, err := New(conf)
			taskManager := &mockTaskManager{}
			tr.SetTaskManager(taskManager)
			So(err, ShouldBeNil)
			So(tr, ShouldNotBeNil)
			tribes = append(tribes, tr)
		}
		var wg sync.WaitGroup
		for _, tr := range tribes {
			timer := time.After(4 * time.Second)
			wg.Add(1)
			go func(tr *tribe) {
				defer wg.Done()
				for {
					select {
					case <-timer:
						panic("timed out establishing membership")
					default:
						if len(tr.members) == len(tribes) {
							return
						}
						logger.Debugf("%v has %v members", tr.memberlist.LocalNode().Name, len(tr.memberlist.Members()))
						time.Sleep(50 * time.Millisecond)
					}
				}
			}(tr)
		}
		wg.Wait()
		Convey("agreements are added", func() {
			t := tribes[rand.Intn(len(tribes))]
			serr := t.AddAgreement(agreement1)
			So(serr, ShouldBeNil)
			err := t.AddPlugin(agreement1, plugin1)
			So(err, ShouldBeNil)
			serr = t.AddTask(agreement1, task1)
			So(serr, ShouldBeNil)
			So(len(t.agreements), ShouldEqual, 1)
			Convey("the state is consistent across the tribe", func() {
				wg = sync.WaitGroup{}
				timedOut := false
				for _, tr := range tribes {
					timer := time.After(10 * time.Second)
					wg.Add(1)
					go func(tr *tribe) {
						defer wg.Done()
						for {
							select {
							case <-timer:
								timedOut = true
								return
							default:
								if a, ok := tr.agreements[agreement1]; ok {
									if a.PluginAgreement != nil {
										if ok, _ := a.PluginAgreement.Plugins.Contains(plugin1); ok {
											return
										}
									}
								}
								logger.Debugf("%v has %v agreements", tr.memberlist.LocalNode().Name, len(tr.agreements))
								time.Sleep(200 * time.Millisecond)
							}
						}
					}(tr)
				}
				wg.Wait()
				So(timedOut, ShouldBeFalse)

				Convey("all members are added to the agreements", func() {
					for _, tr := range tribes {
						logger.Debugf("joining %v %v", agreement1, tr.memberlist.LocalNode().Name)
						err := t.JoinAgreement(agreement1, tr.memberlist.LocalNode().Name)
						So(err, ShouldBeNil)
					}
				})
			})
		})
	})
}

func TestTribeFullStateSyncOnJoin(t *testing.T) {
	numOfTribes := 5
	tribes := []*tribe{}
	agreement1 := "agreement1"
	plugin1 := agreement.Plugin{Name_: "plugin1", Version_: 1}
	plugin2 := agreement.Plugin{Name_: "plugin2", Version_: 1}
	task1 := agreement.Task{ID: uuid.New()}
	task2 := agreement.Task{ID: uuid.New()}
	Convey("A seed is started", t, func() {
		conf := getTestConfig()
		conf.Name = "seed"
		seed, err := New(conf)
		So(seed, ShouldNotBeNil)
		So(err, ShouldBeNil)
		taskManager := &mockTaskManager{}
		seed.SetTaskManager(taskManager)
		tribes = append(tribes, seed)
		Convey("agreements are added", func() {
			seed.AddAgreement(agreement1)
			seed.JoinAgreement(agreement1, seed.memberlist.LocalNode().Name)
			seed.AddPlugin(agreement1, plugin1)
			seed.AddPlugin(agreement1, plugin2)
			seed.AddTask(agreement1, task1)
			seed.AddTask(agreement1, task2)
			So(seed.intentBuffer, ShouldBeEmpty)
			So(len(seed.members), ShouldEqual, 1)
			So(len(seed.members[seed.memberlist.LocalNode().Name].PluginAgreement.Plugins), ShouldEqual, 2)
			So(len(seed.members[seed.memberlist.LocalNode().Name].TaskAgreements[agreement1].Tasks), ShouldEqual, 2)
			Convey("members are added", func() {
				for i := 1; i < numOfTribes; i++ {
					conf := getTestConfig()
					conf.Name = fmt.Sprintf("member-%v", i)
					conf.Seed = fmt.Sprintf("%v:%v", "127.0.0.1", seed.memberlist.LocalNode().Port)
					tr, err := New(conf)
					taskManager := &mockTaskManager{}
					tr.SetTaskManager(taskManager)
					So(err, ShouldBeNil)
					So(tr, ShouldNotBeNil)
					tribes = append(tribes, tr)
				}
				var wg sync.WaitGroup
				timedOut := false
				for _, tr := range tribes {
					timer := time.After(5 * time.Second)
					wg.Add(1)
					go func(tr *tribe) {
						defer wg.Done()
						for {
							select {
							case <-timer:
								timedOut = false
								return
							default:
								if len(tr.memberlist.Members()) == len(tribes) {
									return
								}
							}
						}
					}(tr)
				}
				wg.Wait()
				So(timedOut, ShouldBeFalse)
				for i := 0; i < numOfTribes; i++ {
					log.Debugf("%v is reporting %v members", i, len(tribes[i].memberlist.Members()))
					So(len(tribes[i].memberlist.Members()), ShouldEqual, numOfTribes)
				}
				Convey("members agree on tasks and plugins", func() {
					for _, tr := range tribes {
						So(len(tr.agreements), ShouldEqual, 1)
						So(len(tr.agreements[agreement1].PluginAgreement.Plugins), ShouldEqual, 2)
						So(len(tr.agreements[agreement1].TaskAgreement.Tasks), ShouldEqual, 2)
						So(len(tr.members[seed.memberlist.LocalNode().Name].PluginAgreement.Plugins), ShouldEqual, 2)
						So(len(tr.members[seed.memberlist.LocalNode().Name].TaskAgreements[agreement1].Tasks), ShouldEqual, 2)
					}
					Convey("new members join agreement", func() {
						for _, tr := range tribes {
							if tr.memberlist.LocalNode().Name == "seed" {
								continue
							}
							tr.JoinAgreement(agreement1, tr.memberlist.LocalNode().Name)
						}
						Convey("members agree on tasks, plugins and membership", func() {
							var wg sync.WaitGroup
							timedOut := false
							for _, tr := range tribes {
								wg.Add(1)
								timer := time.After(5 * time.Second)
								go func(tr *tribe) {
									defer wg.Done()
									for {
										select {
										case <-timer:
											timedOut = true
											return
										default:
											if _, ok := tr.members[tr.memberlist.LocalNode().Name]; ok {
												if tr.members[tr.memberlist.LocalNode().Name].PluginAgreement != nil {
													return
												}
											}
										}
									}
								}(tr)
							}
							wg.Wait()
							So(timedOut, ShouldBeFalse)
							for _, tr := range tribes {
								So(len(tr.agreements), ShouldEqual, 1)
								So(len(tr.agreements[agreement1].PluginAgreement.Plugins), ShouldEqual, 2)
								So(len(tr.agreements[agreement1].TaskAgreement.Tasks), ShouldEqual, 2)
								So(len(tr.members[seed.memberlist.LocalNode().Name].PluginAgreement.Plugins), ShouldEqual, 2)
								So(len(tr.members[seed.memberlist.LocalNode().Name].TaskAgreements[agreement1].Tasks), ShouldEqual, 2)
							}
						})
					})
				})
			})
		})
	})
}

func TestTribeTaskAgreements(t *testing.T) {
	numOfTribes := 5
	tribes := getTribes(numOfTribes, nil)
	Convey(fmt.Sprintf("%d tribes are started", numOfTribes), t, func() {
		for i := 0; i < numOfTribes; i++ {
			log.Debugf("%v is reporting %v members", i, len(tribes[i].memberlist.Members()))
			So(len(tribes[0].memberlist.Members()), ShouldEqual, len(tribes[i].memberlist.Members()))
		}
		Convey("the cluster agrees on membership", func() {
			for i := 0; i < numOfTribes; i++ {
				So(
					len(tribes[0].memberlist.Members()),
					ShouldEqual,
					len(tribes[i].memberlist.Members()),
				)
				So(len(tribes[0].members), ShouldEqual, len(tribes[i].members))
			}

			agreementName := "agreement1"
			agreementName2 := "agreement2"
			task1 := agreement.Task{ID: uuid.New()}
			task2 := agreement.Task{ID: uuid.New()}
			Convey("a member handles", func() {
				t := tribes[0]
				t2 := tribes[1]
				Convey("an out of order 'add task' message", func() {
					msg := &taskMsg{
						LTime:         t.clock.Increment(),
						UUID:          uuid.New(),
						TaskID:        task1.ID,
						AgreementName: agreementName,
						Type:          addTaskMsgType,
					}
					b := t.handleAddTask(msg)
					So(b, ShouldEqual, true)
					t.broadcast(addTaskMsgType, msg, nil)
					So(len(t.intentBuffer), ShouldEqual, 1)
					err := t.AddTask(agreementName, task1)
					So(err.Error(), ShouldResemble, errAgreementDoesNotExist.Error())
					err = t.AddAgreement(agreementName)
					So(err, ShouldBeNil)
					So(len(t.intentBuffer), ShouldEqual, 0)
					So(len(t.agreements[agreementName].TaskAgreement.Tasks), ShouldEqual, 1)
					ok, _ := t.agreements[agreementName].TaskAgreement.Tasks.Contains(task1)
					So(ok, ShouldBeTrue)
					Convey("adding an existing task", func() {
						err := t.AddTask(agreementName, task1)
						So(err.Error(), ShouldResemble, errTaskAlreadyExists.Error())
						Convey("removing a task that doesn't exist", func() {
							err := t.RemoveTask(agreementName, agreement.Task{ID: uuid.New()})
							So(err.Error(), ShouldResemble, errTaskDoesNotExist.Error())
							err = t.RemoveTask("doesn't exist", task1)
							So(err.Error(), ShouldResemble, errAgreementDoesNotExist.Error())
							Convey("joining an agreement with tasks", func() {
								err := t.AddAgreement(agreementName2)
								So(err, ShouldBeNil)
								err = t.AddTask(agreementName2, task2)
								So(err, ShouldBeNil)
								err = t.JoinAgreement(agreementName, t.memberlist.LocalNode().Name)
								So(err, ShouldBeNil)
								err = t.JoinAgreement(agreementName, t2.memberlist.LocalNode().Name)
								So(err, ShouldBeNil)
								err = t.JoinAgreement(agreementName2, t.memberlist.LocalNode().Name)
								So(err, ShouldBeNil)
								So(len(t.members[t.memberlist.LocalNode().Name].TaskAgreements), ShouldEqual, 2)
								err = t.canJoinAgreement(agreementName2, t.memberlist.LocalNode().Name)
								So(err, ShouldBeNil)
								So(t.members[t.memberlist.LocalNode().Name].PluginAgreement, ShouldNotBeNil)
								So(len(t.members[t.memberlist.LocalNode().Name].TaskAgreements), ShouldEqual, 2)
								Convey("all members agree on tasks", func(c C) {
									var wg sync.WaitGroup
									for _, t := range tribes {
										wg.Add(1)
										go func(t *tribe) {
											defer wg.Done()
											for {
												if a, ok := t.agreements[agreementName]; ok {
													if ok, _ := a.TaskAgreement.Tasks.Contains(task1); ok {
														return
													}
												}
												time.Sleep(50 * time.Millisecond)
											}
										}(t)
									}
									wg.Wait()
									Convey("the agreement is queried for the state of a given task", func() {
										t := tribes[rand.Intn(numOfTribes)]
										resp := t.taskStateQuery(agreementName, task1.ID)
										So(resp, ShouldNotBeNil)
										responses := taskStateResponses{}
										for r := range resp.resp {
											responses = append(responses, r)
										}
										So(len(responses), ShouldEqual, 2)
										So(responses.State(), ShouldEqual, core.TaskSpinning)
										Convey("a member handles removing a task", func() {
											t := tribes[rand.Intn(numOfTribes)]
											ok, _ := t.agreements[agreementName].TaskAgreement.Tasks.Contains(task1)
											So(ok, ShouldBeTrue)
											err := t.RemoveTask(agreementName, task1)
											ok, _ = t.agreements[agreementName].TaskAgreement.Tasks.Contains(task1)
											So(ok, ShouldBeFalse)
											So(err, ShouldBeNil)
											So(t.intentBuffer, ShouldBeEmpty)
											var wg sync.WaitGroup
											for _, t := range tribes {
												wg.Add(1)
												go func(t *tribe) {
													defer wg.Done()
													for {
														if a, ok := t.agreements[agreementName]; ok {
															if len(a.TaskAgreement.Tasks) == 0 {
																return
															}
														}
														time.Sleep(50 * time.Millisecond)
													}
												}(t)
											}
											wg.Wait()
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

func TestTribeAgreements(t *testing.T) {
	numOfTribes := 5
	tribes := getTribes(numOfTribes, nil)
	Convey(fmt.Sprintf("%d tribes are started", numOfTribes), t, func() {
		for i := 0; i < numOfTribes; i++ {
			log.Debugf("%v is reporting %v members", i, len(tribes[i].memberlist.Members()))
			So(len(tribes[0].memberlist.Members()), ShouldEqual, len(tribes[i].memberlist.Members()))
		}
		Convey("The cluster agrees on membership", func() {
			for i := 0; i < numOfTribes; i++ {

				So(
					len(tribes[0].memberlist.Members()),
					ShouldEqual,
					len(tribes[i].memberlist.Members()),
				)
				So(len(tribes[0].members), ShouldEqual, len(tribes[i].members))
			}

			Convey("A member handles", func() {
				agreementName := "agreement1"
				t := tribes[0]
				t2 := tribes[1]
				Convey("an out-of-order join agreement message", func() {
					msg := &agreementMsg{
						LTime:         t.clock.Increment(),
						UUID:          uuid.New(),
						AgreementName: agreementName,
						MemberName:    t.memberlist.LocalNode().Name,
						Type:          joinAgreementMsgType,
					}
					msg2 := &agreementMsg{
						LTime:         t.clock.Increment(),
						UUID:          uuid.New(),
						AgreementName: agreementName,
						MemberName:    t2.memberlist.LocalNode().Name,
						Type:          joinAgreementMsgType,
					}

					b := t.handleJoinAgreement(msg)
					So(b, ShouldEqual, true)
					So(len(t.intentBuffer), ShouldEqual, 1)
					t.broadcast(joinAgreementMsgType, msg, nil)
					timer := time.After(2 * time.Second)
				loop1:
					for {
						select {
						case <-timer:
							So("Timed out", ShouldEqual, "")
						default:
							if len(t2.intentBuffer) > 0 {
								break loop1
							}
						}
					}
					So(len(t2.intentBuffer), ShouldEqual, 1)

					b = t.handleJoinAgreement(msg2)
					So(b, ShouldEqual, true)
					So(len(t.intentBuffer), ShouldEqual, 2)
					t.broadcast(joinAgreementMsgType, msg2, nil)

					timer = time.After(2 * time.Second)
				loop2:
					for {
						select {
						case <-timer:
							So("Timed out", ShouldEqual, "")
						default:
							if len(t2.intentBuffer) == 2 {
								break loop2
							}
						}
					}
					So(len(t2.intentBuffer), ShouldEqual, 2)

					Convey("an out-of-order add plugin message", func() {
						plugin := agreement.Plugin{Name_: "plugin1", Version_: 1}
						msg := &pluginMsg{
							LTime:         t.clock.Increment(),
							UUID:          uuid.New(),
							Plugin:        plugin,
							AgreementName: agreementName,
							Type:          addPluginMsgType,
						}
						b := t.handleAddPlugin(msg)
						So(b, ShouldEqual, true)
						So(len(t.intentBuffer), ShouldEqual, 3)
						t.broadcast(addPluginMsgType, msg, nil)

						Convey("an add agreement", func() {
							err := t.AddAgreement(agreementName)
							So(err, ShouldBeNil)
							err = t.AddAgreement(agreementName)
							So(err.Error(), ShouldResemble, errAgreementAlreadyExists.Error())
							var wg sync.WaitGroup
							for _, t := range tribes {
								wg.Add(1)
								go func(t *tribe) {
									defer wg.Done()
									for {
										if a, ok := t.agreements[agreementName]; ok {
											logger.Debugf("%s has %d plugins in agreement '%s' and %d intents", t.memberlist.LocalNode().Name, len(t.agreements[agreementName].PluginAgreement.Plugins), agreementName, len(t.intentBuffer))
											if ok, _ := a.PluginAgreement.Plugins.Contains(plugin); ok {
												if len(t.intentBuffer) == 0 {
													return
												}
											}
										}
										logger.Debugf("%s has %d intents", t.memberlist.LocalNode().Name, len(t.intentBuffer))
										time.Sleep(50 * time.Millisecond)
									}
								}(t)
							}
							wg.Wait()

							Convey("being added to an agreement it already belongs to", func() {
								err := t.JoinAgreement(agreementName, t.memberlist.LocalNode().Name)
								So(err.Error(), ShouldResemble, errAlreadyMemberOfPluginAgreement.Error())

								Convey("leaving an agreement that doesn't exist", func() {
									err := t.LeaveAgreement("whatever", t.memberlist.LocalNode().Name)
									So(err.Error(), ShouldResemble, errAgreementDoesNotExist.Error())

									Convey("an unknown member trying to leave an agreement", func() {
										err := t.LeaveAgreement(agreementName, "whatever")
										So(err.Error(), ShouldResemble, errUnknownMember.Error())

										Convey("a member leaving an agreement it isn't part of", func() {
											err := t.LeaveAgreement(agreementName, tribes[2].memberlist.LocalNode().Name)
											So(err, ShouldNotBeNil)
											So(err.Error(), ShouldResemble, errNotAMember.Error())

											Convey("an unknown member trying to join an agreement", func() {
												msg := &agreementMsg{
													LTime:         t.clock.Time(),
													UUID:          uuid.New(),
													AgreementName: agreementName,
													MemberName:    "whatever",
													Type:          joinAgreementMsgType,
												}
												err := t.joinAgreement(msg)
												So(err, ShouldNotBeNil)
												So(err.Error(), ShouldResemble, errUnknownMember.Error())

												Convey("leaving an agreement", func() {
													So(len(t.agreements[agreementName].Members), ShouldEqual, 2)
													So(t.members[t.memberlist.LocalNode().Name].PluginAgreement, ShouldNotBeNil)
													err := t.LeaveAgreement(agreementName, t.memberlist.LocalNode().Name)
													So(err, ShouldBeNil)
													So(len(t.agreements[agreementName].Members), ShouldEqual, 1)
													So(t.members[t.memberlist.LocalNode().Name].PluginAgreement, ShouldBeNil)
													Convey("leaving a tribe results in the member leaving the agreement", func() {
														t2.memberlist.Leave(500 * time.Millisecond)
														timer := time.After(2 * time.Second)
													loop3:
														for {
															select {
															case <-timer:
																So("Timed out", ShouldEqual, "")
															default:
																if len(t.agreements[agreementName].Members) == 0 {
																	break loop3
																}
															}
														}
														So(len(t.agreements[agreementName].Members), ShouldEqual, 0)
														Convey("removes an agreement", func() {
															err := t.RemoveAgreement(agreementName)
															So(err, ShouldBeNil)
															So(len(t.agreements), ShouldEqual, 0)
															Convey("removes an agreement that no longer exists", func() {
																err := t.RemoveAgreement(agreementName)
																So(err.Error(), ShouldResemble, errAgreementDoesNotExist.Error())
																So(len(t.agreements), ShouldEqual, 0)
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
		})
	})
}

func TestTribeMembership(t *testing.T) {
	numOfTribes := 5
	tribes := getTribes(numOfTribes, nil)
	Convey(fmt.Sprintf("%d tribes are started", numOfTribes), t, func() {
		for i := 0; i < numOfTribes; i++ {
			log.Debugf("%v is reporting %v members", i, len(tribes[i].memberlist.Members()))
			So(len(tribes[0].memberlist.Members()), ShouldEqual, len(tribes[i].memberlist.Members()))
		}
		Convey("The cluster agrees on membership", func() {
			for i := 0; i < numOfTribes; i++ {

				So(
					len(tribes[0].memberlist.Members()),
					ShouldEqual,
					len(tribes[i].memberlist.Members()),
				)
				So(len(tribes[0].members), ShouldEqual, len(tribes[i].members))
			}
			Convey("Adds an agreement", func(c C) {
				a := "agreement1"
				t := tribes[numOfTribes-1]
				t.AddAgreement("agreement1")
				var wg sync.WaitGroup
				for _, t := range tribes {
					wg.Add(1)
					go func(t *tribe) {
						defer wg.Done()
						for {
							if t.agreements != nil {
								if _, ok := t.agreements[a]; ok {
									c.So(ok, ShouldEqual, true)
									return
								}
								logger.Debugf(
									"%v has %d agreements",
									t.memberlist.LocalNode().Name,
									len(t.agreements),
								)
							}
							time.Sleep(50 * time.Millisecond)
						}
					}(t)
				}
				wg.Wait()
				for _, t := range tribes {
					So(len(t.agreements), ShouldEqual, 1)
					So(t.agreements[a], ShouldNotBeNil)
				}
				Convey("A member", func() {
					Convey("joins an agreement", func() {
						err := t.JoinAgreement(a, t.memberlist.LocalNode().Name)
						So(err, ShouldBeNil)
					})
					Convey("is added to an agreement it already belongs to", func() {
						Convey("adds a plugin to agreement", func() {
							err := t.AddPlugin(a, agreement.Plugin{Name_: "plugin1", Version_: 1})
							So(err, ShouldBeNil)
						})

						err := t.JoinAgreement(a, t.memberlist.LocalNode().Name)
						So(err.Error(), ShouldResemble, errAlreadyMemberOfPluginAgreement.Error())
					})
					Convey("leaves an agreement that doesn't exist", func() {
						err := t.LeaveAgreement("whatever", t.memberlist.LocalNode().Name)
						So(err.Error(), ShouldResemble, errAgreementDoesNotExist.Error())
					})
					Convey("handles an unknown member trying to leave an agreement", func() {
						err := t.LeaveAgreement(a, "whatever")
						So(err.Error(), ShouldResemble, errUnknownMember.Error())
					})
					Convey("handles a member leaving an agreement it isn't part of", func() {
						err := t.LeaveAgreement(a, tribes[0].memberlist.LocalNode().Name)
						So(err.Error(), ShouldResemble, errNotAMember.Error())
					})
					Convey("handles an unknown member trying to join an agreement", func() {
						msg := &agreementMsg{
							LTime:         t.clock.Time(),
							UUID:          uuid.New(),
							AgreementName: a,
							MemberName:    "whatever",
							Type:          joinAgreementMsgType,
						}
						err := t.joinAgreement(msg)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldResemble, errUnknownMember.Error())
					})
				})
			})
		})

	})
}

func TestTribePluginAgreement(t *testing.T) {
	numOfTribes := 5
	// tribePort := 52600
	tribes := getTribes(numOfTribes, nil)
	Convey(fmt.Sprintf("%d tribes are started", numOfTribes), t, func() {
		for i := 0; i < numOfTribes; i++ {
			So(
				len(tribes[0].memberlist.Members()),
				ShouldEqual,
				len(tribes[i].memberlist.Members()),
			)
			logger.Debugf("%v has %v members", tribes[i].memberlist.LocalNode().Name, len(tribes[i].members))
			So(len(tribes[i].members), ShouldEqual, numOfTribes)
		}

		Convey("The cluster agrees on membership", func() {
			for i := 0; i < numOfTribes; i++ {
				log.Debugf("%v is reporting %v members", i, len(tribes[i].memberlist.Members()))
				So(len(tribes[0].memberlist.Members()), ShouldEqual, len(tribes[i].memberlist.Members()))
				So(len(tribes[0].members), ShouldEqual, len(tribes[i].members))
			}
			oldMember := tribes[0]
			err := tribes[0].memberlist.Leave(2 * time.Second)
			// err := tribes[0].memberlist.Shutdown()
			So(err, ShouldBeNil)
			tribes = append(tribes[:0], tribes[1:]...)

			Convey("Membership decreases as members leave", func(c C) {
				wg := sync.WaitGroup{}
				for i := range tribes {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()
						for {
							if len(tribes[i].members) == len(tribes) {
								c.So(len(tribes[i].members), ShouldEqual, len(tribes))
								return
							}
							time.Sleep(20 * time.Millisecond)
						}
					}(i)
				}
				wg.Wait()
				err := oldMember.memberlist.Shutdown()
				So(err, ShouldBeNil)
				So(len(tribes[rand.Intn(len(tribes))].memberlist.Members()), ShouldEqual, len(tribes))
				So(len(tribes[1].members), ShouldEqual, len(tribes))

				Convey("Membership increases as members join", func(c C) {
					seed := fmt.Sprintf("%v:%v", tribes[0].memberlist.LocalNode().Addr, tribes[0].memberlist.LocalNode().Port)
					conf := getTestConfig()
					conf.Name = fmt.Sprintf("member-%d", numOfTribes+1)
					conf.Seed = seed
					tr, err := New(conf)
					if err != nil {
						So(err, ShouldBeNil)
					}
					tribes = append(tribes, tr)

					wg := sync.WaitGroup{}
					for i := range tribes {
						wg.Add(1)
						go func(i int) {
							defer wg.Done()
							for {
								if len(tribes[i].memberlist.Members()) == len(tribes) {
									c.So(len(tribes[i].members), ShouldEqual, len(tribes))
									return
								}
								time.Sleep(20 * time.Millisecond)
							}
						}(i)
					}
					wg.Wait()
					So(len(tribes[rand.Intn(len(tribes))].memberlist.Members()), ShouldEqual, len(tribes))
					So(len(tribes[rand.Intn(len(tribes))].members), ShouldEqual, len(tribes))

					Convey("Handles a 'add agreement' message broadcasted across the cluster", func(c C) {
						tribes[0].AddAgreement("clan1")
						var wg sync.WaitGroup
						for _, t := range tribes {
							wg.Add(1)
							go func(t *tribe) {
								defer wg.Done()
								for {
									if t.agreements != nil {
										if _, ok := t.agreements["clan1"]; ok {
											c.So(ok, ShouldEqual, true)
											return
										}
									}
									time.Sleep(50 * time.Millisecond)
								}
							}(t)
						}
						wg.Wait()

						numAddMessages := 10
						Convey(fmt.Sprintf("Handles %d plugin 'add messages' broadcasted across the cluster", numAddMessages), func() {
							for i := 0; i < numAddMessages; i++ {
								tribes[0].AddPlugin("clan1", agreement.Plugin{Name_: fmt.Sprintf("plugin%v", i), Version_: 1})
								// time.Sleep(time.Millisecond * 50)
							}
							wg := sync.WaitGroup{}
							for _, tr := range tribes {
								wg.Add(1)
								go func(tr *tribe) {
									defer wg.Done()
									for {
										if clan, ok := tr.agreements["clan1"]; ok {
											if len(clan.PluginAgreement.Plugins) == numAddMessages {
												return
											}
											time.Sleep(50 * time.Millisecond)
											log.Debugf("%v has %v of %v plugins and %d intents\n", tr.memberlist.LocalNode().Name, len(clan.PluginAgreement.Plugins), numAddMessages, len(tr.intentBuffer))
										}
									}
								}(tr)
							}
							log.Debugf("Waits for %d members of clan1 to have %d plugins\n", numOfTribes, numAddMessages)
							wg.Wait()
							for i := 0; i < numOfTribes; i++ {
								So(len(tribes[i].agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, numAddMessages)
								logger.Debugf("%v has %v intents\n", tribes[i].memberlist.LocalNode().Name, len(tribes[i].intentBuffer))
								So(len(tribes[i].intentBuffer), ShouldEqual, 0)
								for k, v := range tribes[i].intentBuffer {
									logger.Debugf("\tadd intent %v %v\n", k, v)

								}
							}

							Convey("Handles duplicate 'add plugin' messages", func() {
								t := tribes[rand.Intn(numOfTribes)]
								msg := &pluginMsg{
									Plugin: agreement.Plugin{
										Name_:    "pluginABC",
										Version_: 1,
									},
									UUID:          uuid.New(),
									AgreementName: "clan1",
									LTime:         t.clock.Time(),
									Type:          addPluginMsgType,
								}
								So(len(t.intentBuffer), ShouldEqual, 0)
								t.handleAddPlugin(msg)
								before := len(t.agreements["clan1"].PluginAgreement.Plugins)
								t.handleAddPlugin(msg)
								after := len(t.agreements["clan1"].PluginAgreement.Plugins)
								So(before, ShouldEqual, after)

								Convey("Handles out-of-order 'add plugin' messages", func() {
									msg := &pluginMsg{
										Plugin: agreement.Plugin{
											Name_:    "pluginABC",
											Version_: 1,
										},
										UUID:          uuid.New(),
										AgreementName: "clan1",
										LTime:         t.clock.Time(),
										Type:          addPluginMsgType,
									}
									t.handleAddPlugin(msg)
									So(len(t.intentBuffer), ShouldEqual, 1)

									Convey("Handles duplicate out-of-order 'add plugin' messages", func() {
										before := len(t.agreements["clan1"].PluginAgreement.Plugins)
										t.handleAddPlugin(msg)
										after := len(t.agreements["clan1"].PluginAgreement.Plugins)
										So(before, ShouldEqual, after)
										So(len(t.intentBuffer), ShouldEqual, 1)
										So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldBeGreaterThan, numAddMessages)
										t.handleRemovePlugin(&pluginMsg{
											LTime:         t.clock.Time(),
											Plugin:        agreement.Plugin{Name_: "pluginABC", Version_: 1},
											AgreementName: "clan1",
											Type:          removePluginMsgType,
										})
										So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldBeGreaterThan, numAddMessages)
										So(len(t.intentBuffer), ShouldEqual, 0)

										// removes the plugin added to test duplicates
										t.handleRemovePlugin(&pluginMsg{
											LTime:         t.clock.Time(),
											Plugin:        agreement.Plugin{Name_: "pluginABC", Version_: 1},
											AgreementName: "clan1",
											Type:          removePluginMsgType,
										})
										So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, numAddMessages)
										// wait for all members of the tribe to get back to 10 plugins
										wg = sync.WaitGroup{}
										for _, tr := range tribes {
											wg.Add(1)
											go func(tr *tribe) {
												defer wg.Done()
												for {
													select {
													case <-time.After(1500 * time.Millisecond):
														c.So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, numAddMessages)
													default:
														if clan, ok := tr.agreements["clan1"]; ok {
															if len(clan.PluginAgreement.Plugins) == numAddMessages {
																return
															}
															time.Sleep(50 * time.Millisecond)
														}
													}
												}
											}(tr)
										}

										Convey("Handles a 'remove plugin' messages broadcasted across the cluster", func(c C) {
											for _, t := range tribes {
												So(len(t.intentBuffer), ShouldEqual, 0)
												So(len(t.intentBuffer), ShouldEqual, 0)
												So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, numAddMessages)
											}
											t := tribes[rand.Intn(numOfTribes)]
											plugin := t.agreements["clan1"].PluginAgreement.Plugins[rand.Intn(numAddMessages)]
											before := len(t.agreements["clan1"].PluginAgreement.Plugins)
											t.RemovePlugin("clan1", plugin)
											after := len(t.agreements["clan1"].PluginAgreement.Plugins)
											So(before-after, ShouldEqual, 1)
											var wg sync.WaitGroup
											for _, t := range tribes {
												wg.Add(1)
												go func(t *tribe) {
													defer wg.Done()
													for {
														select {
														case <-time.After(1500 * time.Millisecond):
															c.So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, after)
														default:
															if len(t.agreements["clan1"].PluginAgreement.Plugins) == after {
																c.So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, after)
																return
															}
															time.Sleep(50 * time.Millisecond)
														}
													}
												}(t)
											}
											wg.Done()

											Convey("Handles out-of-order remove", func() {
												t := tribes[rand.Intn(numOfTribes)]
												plugin := t.agreements["clan1"].PluginAgreement.Plugins[rand.Intn(numAddMessages-1)]
												msg := &pluginMsg{
													LTime:         t.clock.Increment(),
													Plugin:        plugin,
													AgreementName: "clan1",
													UUID:          uuid.New(),
													Type:          removePluginMsgType,
												}
												before := len(t.agreements["clan1"].PluginAgreement.Plugins)
												t.handleRemovePlugin(msg)
												So(before-1, ShouldEqual, len(t.agreements["clan1"].PluginAgreement.Plugins))
												before = len(t.agreements["clan1"].PluginAgreement.Plugins)
												msg.UUID = uuid.New()
												msg.LTime = t.clock.Increment()
												t.handleRemovePlugin(msg)
												after := len(t.agreements["clan1"].PluginAgreement.Plugins)
												So(before, ShouldEqual, after)
												So(len(t.intentBuffer), ShouldEqual, 1)

												Convey("Handles duplicate out-of-order 'remove plugin' messages", func() {
													t.handleRemovePlugin(msg)
													after := len(t.agreements["clan1"].PluginAgreement.Plugins)
													So(before, ShouldEqual, after)
													So(len(t.intentBuffer), ShouldEqual, 1)

													t.handleAddPlugin(&pluginMsg{
														LTime:         t.clock.Increment(),
														Plugin:        plugin,
														AgreementName: "clan1",
														Type:          addPluginMsgType,
													})
													So(len(t.intentBuffer), ShouldEqual, 0)
													ok, _ := t.agreements["clan1"].PluginAgreement.Plugins.Contains(msg.Plugin)
													So(ok, ShouldBeFalse)

													Convey("Handles old 'remove plugin' messages", func() {
														t := tribes[rand.Intn(numOfTribes)]
														plugin := t.agreements["clan1"].PluginAgreement.Plugins[rand.Intn(len(t.agreements["clan1"].PluginAgreement.Plugins))]
														msg := &pluginMsg{
															LTime:         LTime(1025),
															Plugin:        plugin,
															AgreementName: "clan1",
															UUID:          uuid.New(),
															Type:          removePluginMsgType,
														}
														before := len(t.agreements["clan1"].PluginAgreement.Plugins)
														t.handleRemovePlugin(msg)
														after := len(t.agreements["clan1"].PluginAgreement.Plugins)
														So(before-1, ShouldEqual, after)
														msg2 := &pluginMsg{
															LTime:         LTime(513),
															Plugin:        plugin,
															AgreementName: "clan1",
															UUID:          uuid.New(),
															Type:          addPluginMsgType,
														}
														before = len(t.agreements["clan1"].PluginAgreement.Plugins)
														t.handleAddPlugin(msg2)
														after = len(t.agreements["clan1"].PluginAgreement.Plugins)
														So(before, ShouldEqual, after)
														msg3 := &pluginMsg{
															LTime:         LTime(513),
															Plugin:        plugin,
															AgreementName: "clan1",
															UUID:          uuid.New(),
															Type:          removePluginMsgType,
														}
														before = len(t.agreements["clan1"].PluginAgreement.Plugins)
														t.handleRemovePlugin(msg3)
														after = len(t.agreements["clan1"].PluginAgreement.Plugins)
														So(before, ShouldEqual, after)

														Convey("The tribe agrees on plugin agreements", func(c C) {
															var wg sync.WaitGroup
															for _, t := range tribes {
																wg.Add(1)
																go func(t *tribe) {
																	for {
																		defer wg.Done()
																		select {
																		case <-time.After(1 * time.Second):
																			c.So(len(t.memberlist.Members()), ShouldEqual, numOfTribes)
																		default:
																			if len(t.agreements["clan1"].PluginAgreement.Plugins) == numAddMessages-1 {
																				c.So(len(t.agreements["clan1"].PluginAgreement.Plugins), ShouldEqual, numAddMessages-1)
																				return
																			}
																		}

																	}
																}(t)
															}
															wg.Done()
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
	})
}

// seedTribe and conf can be nil
func getTribes(numOfTribes int, seedTribe *tribe) []*tribe {
	tribes := []*tribe{}
	wg := sync.WaitGroup{}
	var seed string
	if seedTribe != nil {
		tribes = append(tribes, seedTribe)
		seed = fmt.Sprintf("%v:%v", seedTribe.memberlist.LocalNode().Addr, seedTribe.memberlist.LocalNode().Port)
	}

	for i := 0; i < numOfTribes; i++ {
		if seedTribe != nil && i == 0 {
			continue
		}
		port := getAvailablePort()
		if seedTribe == nil && i == 0 {
			seed = fmt.Sprintf("127.0.0.1:%d", port)
		}
		// if i > 0 && seedTribe == nil {
		// 	seed = fmt.Sprintf("127.0.0.1:%d", seedPort)
		// }
		conf := GetDefaultConfig()
		conf.Name = fmt.Sprintf("member-%v", i)
		conf.BindAddr = "127.0.0.1"
		conf.BindPort = port
		conf.Seed = seed
		conf.RestAPIPort = getAvailablePort()
		conf.MemberlistConfig.RetransmitMult = conf.MemberlistConfig.RetransmitMult * 2
		tr, err := New(conf)
		taskManager := &mockTaskManager{}
		tr.SetTaskManager(taskManager)
		if err != nil {
			panic(err)
		}
		tribes = append(tribes, tr)
		wg.Add(1)
		to := time.After(15 * time.Second)
		go func(tr *tribe) {
			defer wg.Done()
			for {
				select {
				case <-to:
					panic("Timed out while establishing membership")
				default:
					if len(tr.memberlist.Members()) == numOfTribes {
						return
					}
					log.Debugf("%v has %v of %v members", tr.memberlist.LocalNode().Name, len(tr.memberlist.Members()), numOfTribes)
					members := []string{}
					for _, m := range tr.memberlist.Members() {
						members = append(members, m.Name)
					}
					log.Debugf("%v has %v members", tr.memberlist.LocalNode().Name, members)
					time.Sleep(50 * time.Millisecond)
				}
			}
		}(tr)
	}
	wg.Wait()
	return tribes
}

var nextPort uint64 = 61234

func getAvailablePort() int {
	atomic.AddUint64(&nextPort, 1)
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", nextPort))
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return getAvailablePort()
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
