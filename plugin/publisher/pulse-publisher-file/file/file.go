package file

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin"
)

const (
	name       = "file"
	version    = 1
	pluginType = plugin.PublisherPluginType
)

var (
	defaultPath = "/tmp"
	defaultName = "publish.out"
)

type filePublisher struct {
	name string
	path string
}

func NewFilePublisher() *filePublisher {
	return &filePublisher{
		name: defaultName,
		path: defaultPath,
	}
}

func (f *filePublisher) Publish(metrics []plugin.PluginMetric) error {
	file, err := os.OpenFile(filepath.Join(f.path, f.name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)

	for _, metric := range metrics {
		w.WriteString(fmt.Sprintf("%v|%v|%v\n", time.Now().Format(time.RFC1123Z), strings.Join(metric.Namespace(), "."), metric.Data()))
	}
	w.Flush()
	return nil
}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType)
}
