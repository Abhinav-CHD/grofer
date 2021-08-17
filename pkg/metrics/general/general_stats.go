/*
Copyright © 2020 The PES Open Source Team pesos@pes.edu

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
package general

import (
	"context"
	"sync"

	"github.com/pesos/grofer/pkg/utils"
)

// AggregatedMetrics represents global metrics to be consumed.
type AggregatedMetrics struct {
	NetStats  map[string][]float64
	FieldSet  string
	CpuStats  []float64
	MemStats  []float64
	DiskStats [][]string
	TempStats [][]string
}

type serveFunc func(context.Context, chan AggregatedMetrics) error

// GlobalStats gets stats about the mem and the CPUs and prints it.
func GlobalStats(ctx context.Context, dataChannel chan AggregatedMetrics, refreshRate uint64) error {
	serveFuncs := []serveFunc{
		ServeCPURates,
		ServeMemRates,
		ServeDiskRates,
		ServeNetRates,
		ServeTemperatureRates,
	}

	return utils.TickUntilDone(ctx, refreshRate, func() error {
		var wg sync.WaitGroup

		errCh := make(chan error, len(serveFuncs))

		for _, sf := range serveFuncs {
			wg.Add(1)
			go func(sf serveFunc, dc chan AggregatedMetrics) {
				defer wg.Done()
				errCh <- sf(ctx, dc)
			}(sf, dataChannel)
		}

		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				return err
			}
		}

		return nil
	})
}