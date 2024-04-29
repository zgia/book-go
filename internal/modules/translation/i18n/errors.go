// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package i18n

import (
	"errors"
)

var (
	ErrLocaleAlreadyExist = errors.New("resource already exists")
	ErrUncertainArguments = errors.New("invalid argument")
)
