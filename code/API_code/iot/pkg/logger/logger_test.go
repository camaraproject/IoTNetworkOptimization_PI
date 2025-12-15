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
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Logger_Get(t *testing.T) {
	t.Parallel()

	t.Run("avoid panic when getting logger without initialising config object", func(t *testing.T) {
		Get()
	})

	t.Run("ensure Get returns same instance", func(t *testing.T) {
		first := Get()
		second := Get()
		require.Same(t, first, second, "Expected same logger instance")
	})
}
