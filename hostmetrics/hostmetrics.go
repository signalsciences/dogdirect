package hostmetrics

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// HostMetrics defines a common set of metrics of the host
type HostMetrics struct {
	CPUUser      float64
	CPUSystem    float64
	CPUIowait    float64
	CPUIdle      float64
	CPUStolen    float64
	CPUGuest     float64
	MemTotal     float64
	MemFree      float64
	MemUsed      float64
	MemUsable    float64
	MemPctUsable float64
}

// HostMetricCollector defines book-keeping for the check
type HostMetricCollector struct {
	lastJiffy float64
	lastTimes cpu.TimesStat
}

// NewHostMetricCollector creates a new collector
func NewHostMetricCollector() (*HostMetricCollector, error) {
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("cpu.Times() failed: %s", err)
	} else if len(cpuTimes) == 0 {
		return nil, fmt.Errorf("cpu.Times() returns no cpus")
	}
	t := cpuTimes[0]
	return &HostMetricCollector{
		lastJiffy: t.Total(),
		lastTimes: t,
	}, nil
}

// Run executes the check
func (c *HostMetricCollector) Run() (HostMetrics, error) {
	cpuTimes, err := cpu.Times(false)
	if err != nil {
		// note: can't happen on Linux. gopsutil doesn't
		// return an error
		return HostMetrics{}, fmt.Errorf("cpu.Times() failed: %s", err)
	}
	if len(cpuTimes) == 0 {
		// possible with hardware failure
		return HostMetrics{}, fmt.Errorf("cpu.Times() returns no cpus")
	}
	t := cpuTimes[0]
	jiffy := t.Total()
	toPercent := 100 / (jiffy - c.lastJiffy)

	lastTimes := c.lastTimes
	c.lastJiffy = jiffy
	c.lastTimes = t

	const mbSize float64 = 1024 * 1024
	vmem, err := mem.VirtualMemory()
	if err != nil {
		// only possible if can't parse numbers in /proc/meminfo
		// that would be massive failure
		return HostMetrics{}, fmt.Errorf("mem.VirtualMemory() failed: %s:", err)
	}

	return HostMetrics{
		CPUUser:      ((t.User + t.Nice) - (lastTimes.User + lastTimes.Nice)) * toPercent,
		CPUSystem:    ((t.System + t.Irq + t.Softirq) - (lastTimes.System + lastTimes.Irq + lastTimes.Softirq)) * toPercent,
		CPUIowait:    (t.Iowait - lastTimes.Iowait) * toPercent,
		CPUIdle:      (t.Idle - lastTimes.Idle) * toPercent,
		CPUStolen:    (t.Steal - lastTimes.Steal) * toPercent,
		CPUGuest:     (t.Guest - lastTimes.Guest) * toPercent,
		MemTotal:     float64(vmem.Total) / mbSize,
		MemFree:      float64(vmem.Free) / mbSize,
		MemUsed:      float64(vmem.Total-vmem.Free) / mbSize,
		MemUsable:    float64(vmem.Available) / mbSize,
		MemPctUsable: float64(100-vmem.UsedPercent) / 100,
	}, nil
}
