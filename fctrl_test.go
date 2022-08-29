package fctrl

func generateInput() map[string]interface{} {
	input := make(map[string]interface{})
	input["uid"] = 456
	input["ip"] = "114.19.201.23"
	input["a"] = "this is value a"
	input["b"] = "this is value b"
	return input
}

func mockResourceConfig() *ResourceConfig {
	rule1 := Rule{
		Name:      "10秒不超过4次",
		Period:    "10s",
		Threshold: 4,
		Code:      1001,
		Fields:    []string{"uid", "ip"},
	}
	rule2 := Rule{
		Name:      "15秒不超过8次",
		Period:    "15s",
		Threshold: 8,
		Code:      1002,
		Fields:    []string{"uid", "ip"},
	}
	conf := &ResourceConfig{
		Resource:     "test",
		Rules:        []Rule{rule1, rule2},
		RecordFields: []string{"a", "b"},
	}
	return conf
}
