package availability

type AvailablePlugin struct {
	sub *Subscriptions
}

func (ap *AvailablePlugin) Subscribe() {
	ap.sub.Add()
}

func (ap *AvailablePlugin) Unsubscribe() error {
	err := ap.sub.Remove()
	if err != nil {
		return err
	}
	return nil
}

func (ap *AvailablePlugin) Subscriptions() int {
	return ap.sub.Count()
}
