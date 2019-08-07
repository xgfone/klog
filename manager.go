// Copyright 2019 xgfone
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

package klog

import (
	"sync"
)

// Manager is used to manage a set of the loggers.
type Manager struct {
	lock    sync.RWMutex
	loggers map[string]*Logger
}

// NewManager returns a new Manager, which will add the global logger, Std,
// by default.
func NewManager() *Manager {
	m := &Manager{loggers: make(map[string]*Logger, 8)}
	m.loggers[Std.GetName()] = &Std
	return m
}

// GetLogger returns a Logger named name, which will clone a new one
// with the name by the global Logger, Std, if no this logger.
func (m *Manager) GetLogger(name string) *Logger {
	m.lock.RLock()
	logger, ok := m.loggers[name]
	if !ok {
		log := Std.WithName(name)
		logger = &log
		m.loggers[name] = logger
	}
	m.lock.RUnlock()

	return logger
}

// AddField adds the fields to all the loggers.
func (m *Manager) AddField(fields ...Field) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.AddField(fields...)
	}
	m.lock.Unlock()
}

// AddHook adds the hooks to all the loggers.
func (m *Manager) AddHook(hooks ...Hook) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.AddHook(hooks...)
	}
	m.lock.Unlock()
}

// AddKv adds the key-value to all the loggers.
func (m *Manager) AddKv(key string, value interface{}) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.AddKv(key, value)
	}
	m.lock.Unlock()
}

// SetDepth sets the depth to all the loggers.
func (m *Manager) SetDepth(depth int) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.SetDepth(depth)
	}
	m.lock.Unlock()
}

// SetEncoder sets the encoder to all the loggers.
func (m *Manager) SetEncoder(encoder Encoder) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.SetEncoder(encoder)
	}
	m.lock.Unlock()
}

// SetLevel sets the level to all the loggers.
func (m *Manager) SetLevel(level Level) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.SetLevel(level)
	}
	m.lock.Unlock()
}

// SetWriter sets the writer to all the loggers.
func (m *Manager) SetWriter(w Writer) {
	m.lock.Lock()
	for _, logger := range m.loggers {
		logger.SetWriter(w)
	}
	m.lock.Unlock()
}
