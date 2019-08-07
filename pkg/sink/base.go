package sink

type BaseSink struct {
	KeyToName map[string]string `yaml:"key_to_name"`
}

func (sink *BaseSink) GetKeyToName() map[string]string {
	return sink.KeyToName
}
