package zhuge

import (
	"bytes"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/pyroscope-io/client/upstream"
)

type SampleSession Session

func NewSampleSession(c SessionConfig) (*SampleSession, error) {
	appName, err := mergeTagsWithAppName(c.AppName, c.Tags)
	if err != nil {
		return nil, err
	}

	ps := &SampleSession{
		upstream:      c.Upstream,
		appName:       appName,
		profileTypes:  c.ProfilingTypes,
		disableGCRuns: c.DisableGCRuns,
		sampleRate:    c.SampleRate,
		uploadRate:    c.UploadRate,
		stopCh:        make(chan struct{}),
		logger:        c.Logger,
		cpuBuf:        &bytes.Buffer{},
		memBuf:        &bytes.Buffer{},
	}

	return ps, nil
}

func (ps *SampleSession) isCPUEnabled() bool {
	for _, t := range ps.profileTypes {
		if t == ProfileCPU {
			return true
		}
	}
	return false
}

func (ps *SampleSession) isMemEnabled() bool {
	for _, t := range ps.profileTypes {
		if t == ProfileInuseObjects || t == ProfileAllocObjects || t == ProfileInuseSpace || t == ProfileAllocSpace {
			return true
		}
	}
	return false
}

func (ps *SampleSession) SampleNow() {
	now := time.Now()
	startTime := now.Truncate(ps.uploadRate)
	endTime := startTime.Add(ps.uploadRate)

	if ps.isCPUEnabled() {
		pprof.StartCPUProfile(ps.cpuBuf)
	}

	time.Sleep(ps.uploadRate)

	if ps.isCPUEnabled() {
		pprof.StopCPUProfile()

		ps.upstream.Upload(&upstream.UploadJob{
			Name:            ps.appName,
			StartTime:       startTime,
			EndTime:         endTime,
			SpyName:         "gospy",
			SampleRate:      100,
			Units:           "samples",
			AggregationType: "sum",
			Format:          upstream.FormatPprof,
			Profile:         copyBuf(ps.cpuBuf.Bytes()),
		})
		ps.cpuBuf.Reset()
	}

	if ps.isMemEnabled() {
		currentGCGeneration := numGC()
		// sometimes GC doesn't run within 10 seconds
		//   in such cases we force a GC run
		//   users can disable it with disableGCRuns option
		if currentGCGeneration == ps.lastGCGeneration && !ps.disableGCRuns {
			runtime.GC()
			currentGCGeneration = numGC()
		}
		if currentGCGeneration != ps.lastGCGeneration {
			pprof.WriteHeapProfile(ps.memBuf)
			curMemBytes := copyBuf(ps.memBuf.Bytes())
			ps.memBuf.Reset()
			if ps.memPrevBytes != nil {
				ps.upstream.Upload(&upstream.UploadJob{
					Name:        ps.appName,
					StartTime:   startTime,
					EndTime:     endTime,
					SpyName:     "gospy",
					SampleRate:  100,
					Format:      upstream.FormatPprof,
					Profile:     curMemBytes,
					PrevProfile: ps.memPrevBytes,
				})
			}
			ps.memPrevBytes = curMemBytes
			ps.lastGCGeneration = currentGCGeneration
		}
	}
}
