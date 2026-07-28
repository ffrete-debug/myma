package docker_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
)

// ContainerStats is a single point-in-time resource sample for one container.
//
// CPUPercent is normalised the same way `docker stats` reports it: the share of
// one host CPU consumed, so a container saturating four cores reads ~400%.
// CPUCores records how many cores the sample was taken against so a caller can
// render a 0-100% figure if it prefers.
type ContainerStats struct {
	CPUPercent    float64 `json:"cpu_percent"`
	CPUCores      int     `json:"cpu_cores"`
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	MemoryLimitMB float64 `json:"memory_limit_mb"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRxMB   float64 `json:"network_rx_mb"`
	NetworkTxMB   float64 `json:"network_tx_mb"`
	SampledAt     int64   `json:"sampled_at"`
}

const bytesPerMB = 1024 * 1024

// GetContainerStats returns one resource sample for the named container.
//
// CPU usage is a rate, so it can only be derived by diffing two readings. The
// one-shot stats endpoint leaves PreCPUStats zeroed, which would make the first
// frame's "delta" the container's entire lifetime of CPU time and report a
// nonsense percentage. So we open the streaming endpoint and read two frames:
// the daemon fills the second frame's PreCPUStats from the first, which is
// exactly how `docker stats` computes its figure.
//
// The cost is roughly one daemon tick (~1s) of latency per sample.
func (dm *DockerManager) GetContainerStats(containerName string) (*ContainerStats, error) {
	// The manager's shared context is deliberately not used directly: it has no
	// deadline, and a wedged daemon would otherwise hang this call forever.
	ctx, cancel := context.WithTimeout(dm.ctx, 10*time.Second)
	defer cancel()

	resp, err := dm.client.ContainerStats(ctx, containerName, true)
	if err != nil {
		return nil, fmt.Errorf("read container stats: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	dec := json.NewDecoder(resp.Body)

	var raw container.StatsResponse
	for i := 0; i < 2; i++ {
		if err := dec.Decode(&raw); err != nil {
			// One frame is still useful for memory and network, which are
			// gauges rather than rates; CPU will correctly report 0.
			if i == 1 {
				break
			}
			return nil, fmt.Errorf("decode container stats: %w", err)
		}
	}

	return statsFromRaw(&raw), nil
}

// statsFromRaw converts a daemon stats document into our flattened shape.
// Split out from the transport so the arithmetic is unit-testable without a
// running Docker daemon.
func statsFromRaw(raw *container.StatsResponse) *ContainerStats {
	s := &ContainerStats{
		CPUCores:  int(raw.CPUStats.OnlineCPUs),
		SampledAt: time.Now().Unix(),
	}

	// CPU is a rate and needs a baseline. An unpopulated PreCPUStats (first
	// frame of a stream, or a one-shot response) is NOT a zero baseline:
	// subtracting it would treat the container's whole lifetime of CPU time as
	// one interval and report a wildly inflated figure. Report 0 instead.
	cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage) - float64(raw.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(raw.CPUStats.SystemUsage) - float64(raw.PreCPUStats.SystemUsage)
	hasBaseline := raw.PreCPUStats.SystemUsage > 0
	if hasBaseline && cpuDelta > 0 && systemDelta > 0 {
		cores := float64(raw.CPUStats.OnlineCPUs)
		if cores == 0 {
			// Older daemons leave OnlineCPUs unset and only populate the
			// per-CPU array; fall back to its length, then to a single core.
			cores = float64(len(raw.CPUStats.CPUUsage.PercpuUsage))
			if cores == 0 {
				cores = 1
			}
			s.CPUCores = int(cores)
		}
		s.CPUPercent = (cpuDelta / systemDelta) * cores * 100.0
	}

	// The daemon reports total memory including the page cache. Subtracting the
	// reclaimable cache is what `docker stats` shows and is the number an
	// operator recognises; without it an idle server looks near its limit.
	usage := float64(raw.MemoryStats.Usage)
	if cache, ok := raw.MemoryStats.Stats["inactive_file"]; ok && float64(cache) < usage {
		usage -= float64(cache)
	}
	s.MemoryUsageMB = usage / bytesPerMB
	s.MemoryLimitMB = float64(raw.MemoryStats.Limit) / bytesPerMB
	if raw.MemoryStats.Limit > 0 {
		s.MemoryPercent = usage / float64(raw.MemoryStats.Limit) * 100.0
	}

	for _, n := range raw.Networks {
		s.NetworkRxMB += float64(n.RxBytes) / bytesPerMB
		s.NetworkTxMB += float64(n.TxBytes) / bytesPerMB
	}

	return s
}
