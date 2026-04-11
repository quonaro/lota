package config

type Validator struct {
	AppConfig
}

func GetValidator(config AppConfig) Validator {
	return Validator{config}
}


type Checker interface { 
	
}


