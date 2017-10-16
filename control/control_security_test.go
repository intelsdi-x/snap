// +build medium

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

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

package control

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/gomit"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/snap/control/fixtures"
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/client"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"
	"github.com/intelsdi-x/snap/core/ctypes"
	"github.com/intelsdi-x/snap/plugin/helper"
)

type MockEmitter struct{}

const (
	tlsTestCAFn  = "snaptest-CA"
	tlsTestSrvFn = "snaptest-srv"
	tlsTestCliFn = "snaptest-cli"
)

var tlsTestCA, tlsTestSrv, tlsTestCli string

var testFilesToRemove []string

func (memitter *MockEmitter) Emit(gomit.EventBody) (int, error) { return 0, nil }

type configTLSMock Config

func (m *configTLSMock) export() *Config {
	return (*Config)(m)
}

func (m *configTLSMock) setTLSCertPath(certPath string) *configTLSMock {
	m.TLSCertPath = certPath
	return m
}

func (m *configTLSMock) setTLSKeyPath(keyPath string) *configTLSMock {
	m.TLSKeyPath = keyPath
	return m
}

func TestMain(m *testing.M) {
	setUpTestMain()
	retCode := m.Run()
	tearDownTestMain()
	os.Exit(retCode)
}

func TestSecureCollector(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Having a secure collector", t, func() {
		var ap *availablePlugin
		Convey("framework should establish secure connection", func() {
			security := client.SecurityTLSExtended(tlsTestCli+fixtures.TestCrtFileExt, tlsTestCli+fixtures.TestKeyFileExt, client.SecureClient, []string{tlsTestCA + fixtures.TestCrtFileExt})
			var err error
			ap, err = runPlugin(plugin.Arg{}.
				SetCertPath(tlsTestSrv+fixtures.TestCrtFileExt).
				SetKeyPath(tlsTestSrv+fixtures.TestKeyFileExt).
				SetCACertPaths(tlsTestCA+fixtures.TestCrtFileExt).
				SetTLSEnabled(true), helper.PluginFilePath("snap-plugin-collector-mock2-grpc"),
				security)
			So(err, ShouldBeNil)
			Convey("and valid plugin client should be obtained", func() {
				cli, isCollector := ap.client.(client.PluginCollectorClient)
				So(isCollector, ShouldBeTrue)
				Convey("Ping should not fail", func() {
					err := cli.Ping()
					So(err, ShouldBeNil)
				})
				Convey("GetConfigPolicy should not fail", func() {
					_, err := cli.GetConfigPolicy()
					So(err, ShouldBeNil)
				})
				Convey("GetMetricTypes should not fail", func() {
					cfg := plugin.ConfigType{ConfigDataNode: cdata.NewNode()}
					_, err := cli.GetMetricTypes(cfg)
					So(err, ShouldBeNil)
				})
				Convey("CollectMetrics should not fail", func() {
					_, err := cli.CollectMetrics([]core.Metric{})
					So(err, ShouldBeNil)
				})
			})
			Reset(func() {
				ap.Kill("end-of-test")
			})
		})
	})
}

func TestSecureProcessor(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Having a secure processor", t, func() {
		var ap *availablePlugin
		Convey("framework should establish secure connection", func() {
			security := client.SecurityTLSExtended(tlsTestCli+fixtures.TestCrtFileExt, tlsTestCli+fixtures.TestKeyFileExt, client.SecureClient, []string{tlsTestCA + fixtures.TestCrtFileExt})
			var err error
			ap, err = runPlugin(plugin.Arg{}.
				SetCertPath(tlsTestSrv+fixtures.TestCrtFileExt).
				SetKeyPath(tlsTestSrv+fixtures.TestKeyFileExt).
				SetCACertPaths(tlsTestCA+fixtures.TestCrtFileExt).
				SetTLSEnabled(true), helper.PluginFilePath("snap-plugin-processor-passthru-grpc"),
				security)
			So(err, ShouldBeNil)
			Convey("and valid plugin client should be obtained", func() {
				cli, isProcessor := ap.client.(client.PluginProcessorClient)
				So(isProcessor, ShouldBeTrue)
				Convey("Ping should not fail", func() {
					err := cli.Ping()
					So(err, ShouldBeNil)
				})
				Convey("GetConfigPolicy should not fail", func() {
					_, err := cli.GetConfigPolicy()
					So(err, ShouldBeNil)
				})
				Convey("Process should not fail", func() {
					cfg := map[string]ctypes.ConfigValue{}
					_, err := cli.Process([]core.Metric{}, cfg)
					So(err, ShouldBeNil)
				})
			})
			Reset(func() {
				ap.Kill("end-of-test")
			})
		})
	})
}

func TestSecurePublisher(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Having a secure publisher", t, func() {
		var ap *availablePlugin
		Convey("framework should establish secure connection", func() {
			security := client.SecurityTLSExtended(tlsTestCli+fixtures.TestCrtFileExt, tlsTestCli+fixtures.TestKeyFileExt, client.SecureClient, []string{tlsTestCA + fixtures.TestCrtFileExt})
			var err error
			ap, err = runPlugin(plugin.NewArg(int(log.DebugLevel), false).
				SetCertPath(tlsTestSrv+fixtures.TestCrtFileExt).
				SetKeyPath(tlsTestSrv+fixtures.TestKeyFileExt).
				SetCACertPaths(tlsTestCA+fixtures.TestCrtFileExt).
				SetTLSEnabled(true), helper.PluginFilePath("snap-plugin-publisher-mock-file-grpc"),
				security)
			So(err, ShouldBeNil)
			Convey("and valid plugin client should be obtained", func() {
				cli, isPublisher := ap.client.(client.PluginPublisherClient)
				So(isPublisher, ShouldBeTrue)
				Convey("Ping should not fail", func() {
					err := cli.Ping()
					So(err, ShouldBeNil)
				})
				Convey("GetConfigPolicy should not fail", func() {
					_, err := cli.GetConfigPolicy()
					So(err, ShouldBeNil)
				})
				Convey("Publish should not fail", func() {
					cfg := map[string]ctypes.ConfigValue{}
					tf, err := ioutil.TempFile("", "mock-file-publisher-output")
					if err != nil {
						panic(err)
					}
					testFilesToRemove = append(testFilesToRemove, tf.Name())
					tf.Close()
					cfg["file"] = ctypes.ConfigValueStr{Value: tf.Name()}
					err = cli.Publish([]core.Metric{}, cfg)
					So(err, ShouldBeNil)
				})
			})
			Reset(func() {
				ap.Kill("end-of-test")
			})
		})
	})
}

func TestSecureStreamingCollector(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	Convey("Having a secure streaming collector", t, func() {
		var ap *availablePlugin
		Convey("framework should establish secure connection", func() {
			security := client.SecurityTLSExtended(tlsTestCli+fixtures.TestCrtFileExt, tlsTestCli+fixtures.TestKeyFileExt, client.SecureClient, []string{tlsTestCA + fixtures.TestCrtFileExt})
			var err error
			ap, err = runPlugin(plugin.Arg{}.
				SetCertPath(tlsTestSrv+fixtures.TestCrtFileExt).
				SetKeyPath(tlsTestSrv+fixtures.TestKeyFileExt).
				SetCACertPaths(tlsTestCA+fixtures.TestCrtFileExt).
				SetTLSEnabled(true), helper.PluginFilePath("snap-plugin-streaming-collector-rand1"),
				security)
			So(err, ShouldBeNil)
			Convey("and valid plugin client should be obtained", func() {
				cli, isStreamer := ap.client.(client.PluginStreamCollectorClient)
				So(isStreamer, ShouldBeTrue)
				Convey("Ping should not fail", func() {
					err := cli.Ping()
					So(err, ShouldBeNil)
				})
				Convey("GetConfigPolicy should not fail", func() {
					_, err := cli.GetConfigPolicy()
					So(err, ShouldBeNil)
				})
				Convey("GetMetricTypes should not fail", func() {
					cfg := plugin.ConfigType{ConfigDataNode: cdata.NewNode()}
					_, err := cli.GetMetricTypes(cfg)
					So(err, ShouldBeNil)
				})
				Convey("StreamMetrics should not fail", func() {
					cli.UpdateCollectDuration(time.Second)
					cli.UpdateMetricsBuffer(1)
					mtsin := []core.Metric{}
					m := plugin.MetricType{Namespace_: core.NewNamespace(strings.Fields("a b integer")...)}
					mtsin = append(mtsin, m)
					mch, errch, err := cli.StreamMetrics("test-taskID", mtsin)
					So(err, ShouldBeNil)
					Convey("streaming should deliver metrics rather than error", func() {
						select {
						case mtsout := <-mch:
							So(mtsout, ShouldNotBeNil)
							break
						case err := <-errch:
							t.Fatal(err)
						case <-time.After(5 * time.Second):
							t.Fatal("failed to receive response from stream collector")
						}
					})
				})
				Convey("UpdateCollectedMetrics should not fail", func() {
					err := cli.UpdateCollectedMetrics([]core.Metric{})
					So(err, ShouldBeNil)
				})
				Convey("UpdateCollectDuration should not fail", func() {
					err := cli.UpdateCollectDuration(5 * time.Second)
					So(err, ShouldBeNil)
					err = cli.UpdateCollectDuration(1 * time.Second)
					So(err, ShouldBeNil)
				})
				Convey("UpdateMetricsBuffer should not fail", func() {
					err := cli.UpdateMetricsBuffer(100)
					So(err, ShouldBeNil)
					err = cli.UpdateMetricsBuffer(10)
					So(err, ShouldBeNil)
				})
			})
			Reset(func() {
				ap.Kill("end-of-test")
			})
		})
	})
}

func TestInsecureConfigurationFails(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	tcs := []struct {
		name string
		msg  func(*plugin.Arg, *client.GRPCSecurity, func(string))
	}{
		{
			name: "SecureFrameworkInsecurePlugin_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				// note: server CA certs are used to validate client certs (and vice versa)
				*srv = srv.SetTLSEnabled(false)
				f("Attempting TLS connection between secure framework and insecure plugin")
			},
		},
		{
			name: "InvalidPluginCertKeyPair_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				*srv = srv.SetCertPath(tlsTestCA + fixtures.TestCrtFileExt)
				f("Attempting TLS connection between secure framework and plugin with invalid cert-key pair")
			},
		},
		{
			name: "BadCACertInFramework_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				cli.CACertPaths = []string{tlsTestCA + fixtures.TestBadCrtFileExt}
				f("Attempting TLS connection between framework and plugin using incompatible CA certs, bad cert in framework")
			},
		},
		{
			name: "PluginCACertsUnknownToFramework_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				cli.CACertPaths = []string{}
				f("Attempting TLS connection between secure framework and plugin with certificate without CA certs known to framework")
			},
		},
		{
			name: "InsecureFrameworkSecurePlugin_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				cli.TLSEnabled = false
				f("Attempting TLS connection between insecure framework and secure plugin")
			},
		},
		{
			name: "InvalidFrameworkCertKeyPair_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				cli.TLSCertPath = tlsTestCA + fixtures.TestCrtFileExt
				f("Attempting TLS connection between invalid framework with invalid cert-key pair and secure plugin")
			},
		},
		{
			name: "FrameworkCACertsUnknownToPlugin_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				srv.CACertPaths = ""
				f("Attempting TLS connection between invalid framework with no CA certs known to plugin and secure plugin")
			},
		},
		{
			name: "BadCACertInPlugin_Fail",
			msg: func(srv *plugin.Arg, cli *client.GRPCSecurity, f func(string)) {
				srv.CACertPaths = tlsTestCA + fixtures.TestBadCrtFileExt
				f("Attempting TLS connection between framework and plugin using incompatible CA certs, bad cert in plugin")
			},
		},
	}
	for _, tc := range tcs {
		security := client.SecurityTLSExtended(tlsTestCli+fixtures.TestCrtFileExt, tlsTestCli+fixtures.TestKeyFileExt, client.SecureClient, []string{tlsTestCA + fixtures.TestCrtFileExt})
		pluginArgs := plugin.Arg{}.
			SetCertPath(tlsTestSrv + fixtures.TestCrtFileExt).
			SetKeyPath(tlsTestSrv + fixtures.TestKeyFileExt).
			SetCACertPaths(tlsTestCA + fixtures.TestCrtFileExt).
			SetTLSEnabled(true)
		runThisCase := func(f func(msg string)) {
			t.Run(tc.name, func(_ *testing.T) {
				tc.msg(&pluginArgs, &security, f)
			})
		}
		runThisCase(func(msg string) {
			Convey(msg, t, func() {
				var ap *availablePlugin
				Convey("should fail", func() {
					So(func() {
						var err error
						ap, err = runPlugin(pluginArgs, helper.PluginFilePath("snap-plugin-collector-mock2-grpc"),
							security)
						// currently grpc may not return error immediately; attempt to ping
						if err != nil {
							panic(err)
						}
						cli, isCollector := ap.client.(client.PluginCollectorClient)
						So(isCollector, ShouldBeTrue)
						err = cli.Ping()
						if err != nil {
							panic(err)
						}
					}, ShouldPanic)
				})
				Reset(func() {
					if ap != nil {
						ap.Kill("end-of-test")
					}
				})
			})
		})
	}
}

func (m *configTLSMock) setCACertPaths(caCertPaths string) *configTLSMock {
	m.CACertPaths = caCertPaths
	return m
}

func TestSecuritySetupFromConfig(t *testing.T) {
	var (
		fakeSampleCert         = "/fake-samples/certs/server-cert"
		fakeSampleKey          = "/fake-samples/keys/server-key"
		fakeSampleCACertsSplit = []string{"/fake-samples/root-ca/ca-one", "/fake-samples/root-ca/ca-two"}
		fakeSampleCACerts      = strings.Join(fakeSampleCACertsSplit, string(filepath.ListSeparator))
	)
	tcs := []struct {
		name           string
		msg            func(func(string))
		cfg            *Config
		wantError      bool
		wantRunnersec  client.GRPCSecurity
		wantManagersec client.GRPCSecurity
	}{
		{
			name: "DefaultEmptyConfig",
			msg: func(f func(string)) {
				f("passing default (empty) config values, initialization should succeed and result in security disabled")
			},
			cfg:            GetDefaultConfig(),
			wantError:      false,
			wantRunnersec:  client.SecurityTLSOff(),
			wantManagersec: client.SecurityTLSOff(),
		},
		{
			name: "TLSEnabledForwardedToSubmodules",
			msg: func(f func(string)) {
				f("having TLS enabled in config, plugin runner and manager receive same security values")
			},
			cfg: (*configTLSMock)(GetDefaultConfig()).
				setTLSCertPath(fakeSampleCert).
				setTLSKeyPath(fakeSampleKey).
				export(),
			wantError:      false,
			wantRunnersec:  client.SecurityTLSEnabled(fakeSampleCert, fakeSampleKey, client.SecureClient),
			wantManagersec: client.SecurityTLSEnabled(fakeSampleCert, fakeSampleKey, client.SecureClient),
		},
		{
			name: "TLSEnabledCACertsForwardedToSubmodules",
			msg: func(f func(string)) {
				f("having TLS enabled with CA cert paths in config, plugin runner and manager receive same security values")
			},
			cfg: (*configTLSMock)(GetDefaultConfig()).
				setTLSCertPath(fakeSampleCert).
				setTLSKeyPath(fakeSampleKey).
				setCACertPaths(fakeSampleCACerts).
				export(),
			wantError:      false,
			wantRunnersec:  client.SecurityTLSExtended(fakeSampleCert, fakeSampleKey, client.SecureClient, fakeSampleCACertsSplit),
			wantManagersec: client.SecurityTLSExtended(fakeSampleCert, fakeSampleKey, client.SecureClient, fakeSampleCACertsSplit),
		},
	}
	var gotRunner *runner
	var gotManager *pluginManager

	for _, tc := range tcs {
		oldRunnerOpts, oldManagerOpts := append([]pluginRunnerOpt{}, defaultRunnerOpts...), append([]pluginManagerOpt{}, defaultManagerOpts...)
		defaultRunnerOpts = append(defaultRunnerOpts, func(r *runner) {
			gotRunner = r
		})
		defaultManagerOpts = append(defaultManagerOpts, func(m *pluginManager) {
			gotManager = m
		})
		runThisCase := func(f func(msg string)) {
			t.Run(tc.name, func(_ *testing.T) {
				Convey("Initializing plugin control module", t, func() {
					tc.msg(f)
				})
			})
		}
		runThisCase(func(msg string) {
			Convey(msg, func() {
				if tc.wantError {
					So(func() {
						New(tc.cfg)
					}, ShouldPanic)
					return
				}
				So(func() {
					New(tc.cfg)
				}, ShouldNotPanic)
				So(gotRunner.grpcSecurity, ShouldResemble, tc.wantRunnersec)
				So(gotManager.grpcSecurity, ShouldResemble, tc.wantManagersec)
			})
			Reset(func() {
				defaultRunnerOpts, defaultManagerOpts = oldRunnerOpts, oldManagerOpts
			})
		})
	}
}

func runPlugin(args plugin.Arg, pluginPath string, security client.GRPCSecurity) (*availablePlugin, error) {
	ep, err := plugin.NewExecutablePlugin(args, pluginPath)
	if err != nil {
		panic(err)
	}
	var r *runner
	if security.TLSEnabled {
		r = newRunner(OptEnableRunnerTLS(security))
	} else {
		r = newRunner()
	}
	r.SetEmitter(new(MockEmitter))
	ap, err := r.startPlugin(ep)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func setUpTestMain() {
	rand.Seed(time.Now().Unix())
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("unable to reach current dir for generating TLS certificates: %v", err))
	}
	u := fixtures.CertTestUtil{Prefix: cwd}
	if tlsTestFiles, err := u.StoreTLSCerts(tlsTestCAFn, tlsTestSrvFn, tlsTestCliFn); err != nil {
		panic(err)
	} else {
		testFilesToRemove = append(testFilesToRemove, tlsTestFiles...)
	}
	tlsTestCA = filepath.Join(cwd, tlsTestCAFn)
	tlsTestSrv = filepath.Join(cwd, tlsTestSrvFn)
	tlsTestCli = filepath.Join(cwd, tlsTestCliFn)
}

func tearDownTestMain() {
	for _, fn := range testFilesToRemove {
		os.Remove(fn)
	}
}
