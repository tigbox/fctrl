package fctrl

type ProductRules struct {
	ProductName  string   `toml:"product_name" yaml:"product_name" json:"product_name"`
	Rules        []Rule   `toml:"rules" yaml:"rules" json:"rules"`
	RecordFields []string `toml:"record_fields" yaml:"record_fields" json:"record_fields"`
}

type Rule struct {
	Name      string   `toml:"name" yaml:"name" json:"name"`
	Period    string   `toml:"period" yaml:"period" json:"period"`
	Threshold uint64   `toml:"threshold" yaml:"threshold" json:"threshold"`
	Code      int64    `toml:"code" yaml:"code" json:"code"`
	Fields    []string `toml:"fields" yaml:"fields" json:"fields"`
}
