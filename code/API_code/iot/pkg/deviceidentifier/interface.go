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
package deviceidentifier

import (
	"context"

	"github.com/camaraproject/IoTNetworkOptimization_PI/code/API_code/iot/api/models"
)

// Translator converts various device identifiers to networkAccessIdentifier.
type Translator interface {
	// ResolveNetworkAccessIdentifier resolves a device's identifiers to a networkAccessIdentifier.
	ResolveNetworkAccessIdentifier(ctx context.Context, device models.Device) (models.NetworkAccessIdentifier, error)
}
