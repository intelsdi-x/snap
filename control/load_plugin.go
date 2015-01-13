// Load_plugin.go implements load, unload, and swap behavior for plugins.
package control

import (
	"errors"
)

func (p *pluginControl) UnloadPlugin(lPlugin *LoadedPlugin) error {
	if !p.Started {
		return errors.New("Must start plugin control before calling Load()")
	}

	if lPlugin.State != LoadedState {
		return errors.New("Plugin must be in a LoadedState")
	}

	var pluginRemoved = false
	for i, lp := range p.LoadedPlugins {
		if lp == lPlugin {
			pluginRemoved = true
			p.LoadedPlugins = append(p.LoadedPlugins[:i], p.LoadedPlugins[i+1:]...)
		}
	}

	if !pluginRemoved {
		return errors.New("Must load plugin before calling Unload()")
	}

	return nil
}
