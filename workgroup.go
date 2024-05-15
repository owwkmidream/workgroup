// Copyright © 2017 Heptio
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

// Package workgroup 为控制一组相关程序的生命周期提供了一种机制。
package workgroup

import "sync"

// Group 一个管理一组具有相关生命周期的程序。
// Group 的零值无需初始化即可完全使用。
type Group struct {
	fn []func(<-chan struct{}) error
}

// Add 将一个函数添加到组中。
// 调用 Run 时，该函数将在自己的 goroutine 中执行。
// Add 必须在 Run 之前调用。
func (g *Group) Add(fn func(<-chan struct{}) error) {
	g.fn = append(g.fn, fn)
}

// Run 会在自己的 goroutine 中执行通过 Add 注册的每个函数。
// Run 阻塞，直到所有函数都返回。
// 第一个返回的函数将触发传递给每个函数的通道的关闭，这些函数也应依次返回。
// 第一个退出的函数的返回值将返回给 Run 的调用者。
func (g *Group) Run() error {
	// 如果没有注册函数，则立即返回。
	if len(g.fn) < 1 {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(len(g.fn))

	stop := make(chan struct{})
	result := make(chan error, len(g.fn))
	for _, fn := range g.fn {
		go func(fn func(<-chan struct{}) error) {
			defer wg.Done()
			result <- fn(stop)
		}(fn)
	}

	defer wg.Wait()
	defer close(stop)
	return <-result
}
