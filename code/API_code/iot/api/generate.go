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
package api

//go:generate go tool oapi-codegen -config config/models.yaml ./iot-network-optimization.yaml
//go:generate go tool oapi-codegen -config config/server.yaml ./iot-network-optimization.yaml
//go:generate go tool oapi-codegen -config config/client.yaml ./iot-network-optimization.yaml
