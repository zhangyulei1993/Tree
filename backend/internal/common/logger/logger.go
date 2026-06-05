package logger

import "go.uber.org/zap"

func New(level string, env string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "prod" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	if level != "" {
		if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
			return nil, err
		}
	}

	return cfg.Build()
}

func Error(err error) zap.Field {
	return zap.Error(err)
}
