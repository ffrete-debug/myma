package docker_manager

import (
	"math"
	"testing"

	"github.com/docker/docker/api/types/container"
)

func approx(t *testing.T, got, want float64, label string) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

// A container using half of each of 4 cores should read ~200%, matching how
// `docker stats` reports multi-core usage.
func TestStatsFromRawCPUPercentIsCoreScaled(t *testing.T) {
	raw := &container.StatsResponse{}
	raw.CPUStats.CPUUsage.TotalUsage = 5_000_000
	raw.CPUStats.SystemUsage = 10_000_000
	raw.CPUStats.OnlineCPUs = 4
	raw.PreCPUStats.CPUUsage.TotalUsage = 4_000_000
	raw.PreCPUStats.SystemUsage = 8_000_000

	// cpuDelta 1e6 / systemDelta 2e6 = 0.5, x4 cores x100 = 200%
	approx(t, statsFromRaw(raw).CPUPercent, 200.0, "CPUPercent")
}

// A just-started container has no previous sample. The deltas are zero and the
// result must be 0, not NaN from dividing by zero.
func TestStatsFromRawNoPreviousSample(t *testing.T) {
	raw := &container.StatsResponse{}
	raw.CPUStats.CPUUsage.TotalUsage = 5_000_000
	raw.CPUStats.SystemUsage = 10_000_000
	raw.CPUStats.OnlineCPUs = 2

	got := statsFromRaw(raw).CPUPercent
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Fatalf("CPUPercent = %v, want a finite 0", got)
	}
	approx(t, got, 0.0, "CPUPercent")
}

// Older daemons leave OnlineCPUs unset and only fill the per-CPU array.
func TestStatsFromRawFallsBackToPercpuLength(t *testing.T) {
	raw := &container.StatsResponse{}
	raw.CPUStats.CPUUsage.TotalUsage = 2_000_000
	raw.CPUStats.SystemUsage = 4_000_000
	raw.CPUStats.CPUUsage.PercpuUsage = []uint64{1, 2, 3, 4, 5, 6, 7, 8}
	raw.PreCPUStats.CPUUsage.TotalUsage = 1_000_000
	raw.PreCPUStats.SystemUsage = 2_000_000

	s := statsFromRaw(raw)
	if s.CPUCores != 8 {
		t.Errorf("CPUCores = %d, want 8 from PercpuUsage length", s.CPUCores)
	}
	// 1e6/2e6 = 0.5, x8 x100 = 400%
	approx(t, s.CPUPercent, 400.0, "CPUPercent")
}

// Memory must exclude the reclaimable page cache, otherwise an idle ARK server
// reads as nearly out of memory.
func TestStatsFromRawSubtractsPageCache(t *testing.T) {
	raw := &container.StatsResponse{}
	raw.MemoryStats.Usage = 800 * bytesPerMB
	raw.MemoryStats.Limit = 1000 * bytesPerMB
	raw.MemoryStats.Stats = map[string]uint64{"inactive_file": 300 * bytesPerMB}

	s := statsFromRaw(raw)
	approx(t, s.MemoryUsageMB, 500.0, "MemoryUsageMB")
	approx(t, s.MemoryLimitMB, 1000.0, "MemoryLimitMB")
	approx(t, s.MemoryPercent, 50.0, "MemoryPercent")
}

// An unlimited container reports Limit 0; percent must stay 0 rather than
// dividing by zero.
func TestStatsFromRawZeroMemoryLimit(t *testing.T) {
	raw := &container.StatsResponse{}
	raw.MemoryStats.Usage = 128 * bytesPerMB

	s := statsFromRaw(raw)
	if math.IsNaN(s.MemoryPercent) || math.IsInf(s.MemoryPercent, 0) {
		t.Fatalf("MemoryPercent = %v, want a finite 0", s.MemoryPercent)
	}
	approx(t, s.MemoryPercent, 0.0, "MemoryPercent")
	approx(t, s.MemoryUsageMB, 128.0, "MemoryUsageMB")
}

func TestStatsFromRawSumsAllNetworkInterfaces(t *testing.T) {
	raw := &container.StatsResponse{
		Networks: map[string]container.NetworkStats{
			"eth0": {RxBytes: 10 * bytesPerMB, TxBytes: 5 * bytesPerMB},
			"eth1": {RxBytes: 2 * bytesPerMB, TxBytes: 1 * bytesPerMB},
		},
	}

	s := statsFromRaw(raw)
	approx(t, s.NetworkRxMB, 12.0, "NetworkRxMB")
	approx(t, s.NetworkTxMB, 6.0, "NetworkTxMB")
}
