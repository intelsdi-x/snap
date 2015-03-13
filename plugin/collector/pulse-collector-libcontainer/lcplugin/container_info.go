package lcplugin

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/libcontainer"
)

func getContainerInfo(containerPath string) (*libcontainer.Config, *libcontainer.State, *libcontainer.ContainerStats, error) {
	state, err := getState(filepath.Join(containerPath, "state.json"))
	if err != nil {
		log.Printf("Libcontainer: error %s when accessing path %s\n",
			err.Error(), filepath.Join(containerPath, "state.json"))
		return nil, nil, nil, err
	}

	config, err := getConfig(filepath.Join(containerPath, "container.json"))
	if err != nil {
		log.Printf("Libcontainer: error %s when accessing path %s\n",
			err.Error(), filepath.Join(containerPath, "container.json"))
		return nil, nil, nil, err
	}

	stats, err := libcontainer.GetStats(config, state)
	if err != nil {
		log.Printf("Libcontainer: error while GetStats: %s\n", err.Error())
		return nil, nil, nil, err
	}

	return config, state, stats, nil
}

func getState(state_config string) (*libcontainer.State, error) {
	f, err := os.Open(state_config)
	if err != nil {
		log.Printf("failed to open %s - %s\n", state_config, err)
		return nil, err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	retstate := new(libcontainer.State)
	d.Decode(retstate)
	return retstate, nil
}

func getConfig(container_config string) (*libcontainer.Config, error) {
	f, err := os.Open(container_config)
	if err != nil {
		log.Printf("failed to open %s - %s\n", container_config, err)
		return nil, err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	retConfig := new(libcontainer.Config)
	d.Decode(retConfig)

	return retConfig, nil
}
