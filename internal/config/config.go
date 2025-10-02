package config 

type config struct {
	Yaml *yamlConfig
	Flags *flagConfig
}

func NewConfig(pathYaml string) *config {

	ymlCfg := newYamlConfig(pathYaml);
	flag := newFlagConfig();

	return &config{
		Yaml: ymlCfg,
		Flags: flag,
	}
}
