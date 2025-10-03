package config 

type config struct {
	Flags *flagConfig
	Yaml *yamlConfig
}

func NewConfig() *config {

	flag := newFlagConfig();
	ymlCfg := newYamlConfig(flag.Config);
	

	return &config{
		Flags: flag,
		Yaml: ymlCfg,
	}
}
