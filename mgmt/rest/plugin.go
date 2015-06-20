package rest

import (
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
)

type loadedPlugin struct {
	*plugin
	TypeName        string `json:"type"`
	Status          string `json:"status"`
	LoadedTimestamp int64  `json:"loaded_timestamp"`
}

type availablePlugin struct {
	*plugin
	TypeName string    `json:"type"`
	HitCount int       `json:"hit_count"`
	LastHit  time.Time `json:"last_hit"`
	ID       int       `json:"ID"`
}

type plugin struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

func (s *Server) loadPlugin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
		replyError(500, w, err)
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				replySuccess(200, w, nil)
				return
			}
			if err != nil {
				log.Fatal(err)
				replyError(500, w, err)
				return
			}
			b, err := ioutil.ReadAll(p)
			if err != nil {
				replyError(500, w, err)
				return
			}
			// persist plugin to autodiscover path if available or temp location
			autoPaths := s.mm.GetAutodiscoverPaths()
			var f *os.File
			if len(autoPaths) > 0 {
				// write to first autoPath
				// todo preserve the name of the file the user provided
				//f, err = ioutil.TempFile(autoPaths[0], "pulse-plugin")
				f, err = os.Create(path.Join(autoPaths[0], p.FileName()))
			} else {
				// write to temp location
				f, err = os.Create(path.Join(os.TempDir(), p.FileName()))
			}
			if err != nil {
				replyError(500, w, err)
				return
			}
			n, err := f.Write(b)
			log.Debugf("wrote %v to %v", n, f.Name())
			if err != nil {
				replyError(500, w, err)
				return
			}
			err = f.Chmod(0700)
			if err != nil {
				replyError(500, w, err)
				return
			}

			err = s.mm.Load(f.Name())
			if err != nil {
				replyError(500, w, err)
				return
			}
		}
	}
}

func (s *Server) getPlugins(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var detail bool
	for k, _ := range r.URL.Query() {
		if k == "details" {
			detail = true
		}
	}

	data := make(map[string]interface{})
	lps := make([]loadedPlugin, len(s.mm.PluginCatalog()))
	for i, p := range s.mm.PluginCatalog() {
		lps[i] = loadedPlugin{
			plugin: &plugin{
				Name:    p.Name(),
				Version: p.Version(),
			},
			TypeName:        p.TypeName(),
			Status:          p.Status(),
			LoadedTimestamp: p.LoadedTimestamp(),
		}
	}
	data["LoadedPlugins"] = lps

	if detail {
		a := s.mm.AvailablePlugins()
		aps := make([]availablePlugin, len(a))
		for i, p := range a {
			aps[i] = availablePlugin{
				plugin: &plugin{
					Name:    p.Name(),
					Version: p.Version(),
				},
				TypeName: p.TypeName(),
				HitCount: p.HitCount(),
				LastHit:  p.LastHit(),
				ID:       p.ID(),
			}
		}
		data["RunningPlugins"] = aps
	}

	replySuccess(200, w, data)
}

func (s *Server) getPluginsByName(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}

func (s *Server) getPlugin(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
}
