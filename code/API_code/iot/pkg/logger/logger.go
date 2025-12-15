/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package logger

import (
	"log"
	"os"
	"runtime/debug"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/pkg/config"
)

var once sync.Once

var logger *zap.Logger

// Get returns a zap.Logger instance.
func Get() *zap.Logger {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		logConf := config.GetLogConfig()

		level := zap.InfoLevel
		levelEnv := logConf.Level
		if levelEnv != "" {
			levelFromEnv, err := zapcore.ParseLevel(levelEnv)
			if err != nil {
				log.Printf("Invalid log level '%s', defaulting to INFO: %v", levelEnv, err)
			} else {
				level = levelFromEnv
			}
		}

		logLevel := zap.NewAtomicLevelAt(level)

		var encoder zapcore.Encoder
		if logConf.Format == "development" {
			developmentCfg := zap.NewDevelopmentEncoderConfig()
			developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			encoder = zapcore.NewConsoleEncoder(developmentCfg)
		} else {
			productionCfg := zap.NewProductionEncoderConfig()
			productionCfg.TimeKey = "timestamp"
			productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			encoder = zapcore.NewJSONEncoder(productionCfg)
		}

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		logger = zap.New(zapcore.NewCore(encoder, stdout, logLevel).With(
			[]zapcore.Field{
				zap.String("git_revision", gitRevision),
				zap.String("go_version", buildInfo.GoVersion),
			},
		))
	})

	return logger
}

func IsDebug() bool {
	return config.GetLogConfig().Level == "debug"
}
