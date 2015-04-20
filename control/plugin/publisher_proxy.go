package plugin

type PublishArgs struct {
	PluginMetrics []PluginMetric
}

type PublishReply struct {
}

type publisherPluginProxy struct {
	Plugin  PublisherPlugin
	Session Session
}

func (p *publisherPluginProxy) Publish(args PublishArgs, reply *PublishReply) error {
	p.Session.Logger().Println("Publish called")
	p.Session.ResetHeartbeat()

	// TODO refactor to new contract
	// err := p.Plugin.Publish(args.PluginMetrics)
	// if err != nil {
	// return errors.New(fmt.Sprintf("Publish call error: %v", err.Error()))
	// }
	return nil
}
