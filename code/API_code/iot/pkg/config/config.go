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
package config

import (
	"fmt"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Log struct {
	Level  string `split_words:"true" default:"info"`
	Format string `split_words:"true" default:"production"`
}

type API struct {
	Address string `split_words:"true" default:"0.0.0.0:8080"`
}

type Database struct {
	Uri  string `split_words:"true" default:"mongodb://localhost:27017"`
	Name string `split_words:"true" default:"iot"`
}

// HTTP client configuration
type HTTP struct {
	InsecureSkipVerify bool `split_words:"true" default:"false" description:"If true, skip TLS certificate verification for internal cluster services."`
}

type EasyAPI struct {
	BaseURL string `split_words:"true" default:""`
}

type PowerSaving struct {
	MaxLatency      string `split_words:"true" default:"1"`
	MaxResponseTime string `split_words:"true" default:"1"`
}

type Retention struct {
	// Period is the retention duration for completed/failed transactions.
	Period string `split_words:"true" default:"168h"`
	// CleanupInterval is how often the cleanup job runs.
	CleanupInterval string `split_words:"true" default:"1h"`
}

type Config struct {
	API
	Database
	EasyAPI
	HTTP
	PowerSaving
	Retention
	Log
}

func process(prefix string, spec interface{}) {
	if err := envconfig.Process(prefix, spec); err != nil {
		fmt.Printf("failed to load %s config: %v\n", prefix, err)
	}
}

func GetConf() Config {
	var api API
	process("api", &api)

	var db Database
	process("db", &db)

	var easyAPI EasyAPI
	process("easyapi", &easyAPI)

	var powerSaving PowerSaving
	process("powersaving", &powerSaving)

	var retention Retention
	process("retention", &retention)

	var log Log
	process("log", &log)

	var http HTTP
	process("http", &http)

	return Config{api, db, easyAPI, http, powerSaving, retention, log}
}

var (
	logConfig     Log
	loadLogConfig sync.Once
)

func GetLogConfig() Log {
	loadLogConfig.Do(func() {
		logConfig = Log{}
		process("log", &logConfig)
	})
	return logConfig
}
