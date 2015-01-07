package availability

// Handles events pertaining to plugins and control the runnning state accordingly.
type Runner struct {
}

// Runner - Register Handlers (entry point to wire Runner to events)

// Runner - Start (begin handling)

// Runner - Stop (stop handling, gracefully stop all plugins)

// Start a RunnablePlugin returning details on the RunningState
func StartPlugin() {}

// Halt a RunnablePlugin
func StopPlugin() {}
