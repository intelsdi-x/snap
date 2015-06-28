package client

// Functional tests through client to REST API

import (
	"fmt"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/pulse/control"
	"github.com/intelsdi-x/pulse/mgmt/rest"
	"github.com/intelsdi-x/pulse/scheduler"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	NextPort = 9000
)

func getPort() int {
	defer incrPort()
	return NextPort
}

func incrPort() {
	NextPort += 10
}

// REST API instances that are started are killed when the tests end.
// When we eventually have a REST API Stop command this can be killed.
func startAPI(port int) {
	// Start a REST API to talk to
	log.SetLevel(log.FatalLevel)
	r := rest.New()
	c := control.New()
	c.Start()
	s := scheduler.New()
	s.SetMetricManager(c)
	s.Start()
	r.BindMetricManager(c)
	r.BindTaskManager(s)
	r.Start(":" + fmt.Sprint(port))
	time.Sleep(time.Millisecond * 100)
}

func TestPulseClient(t *testing.T) {
	Convey("REST API functional V1", t, func() {
	})
}
