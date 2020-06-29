// Copyright 2020, Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build windows

package main

import (
	"github.com/pkg/errors"
	"go.opentelemetry.io/collector/service"
	"golang.org/x/sys/windows/svc"
)

func run(params service.Parameters) error {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		return errors.Wrap(err, "failed to determine if we are running in an interactive session")
	}

	if isInteractive {
		return runInteractive(params)
	} else {
		return runService(params)
	}
}

func runService(params service.Parameters) error {
	// do not need to supply service name when startup is invoked through Service Control Manager directly
	if err := svc.Run("", service.NewWindowsService(params)); err != nil {
		return errors.Wrap(err, "failed to start service")
	}

	return nil
}
