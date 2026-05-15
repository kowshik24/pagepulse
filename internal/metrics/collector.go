package metrics

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type Point struct {
	TS  time.Time `json:"ts"`
	Val float64   `json:"val"`
}

type Summary struct {
	Timestamp time.Time     `json:"timestamp"`
	Hostname  string        `json:"hostname"`
	OS        string        `json:"os"`
	Arch      string        `json:"arch"`
	UptimeSec uint64        `json:"uptime_sec"`
	CPU       CPUStats      `json:"cpu"`
	Memory    MemoryStats   `json:"memory"`
	Disk      DiskAggregate `json:"disk"`
	Network   NetAggregate  `json:"network"`
	Trends    Trends        `json:"trends"`
}

type CPUStats struct {
	UsagePct float64 `json:"usage_pct"`
}

type MemoryStats struct {
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsedPct    float64 `json:"used_pct"`
}

type DiskAggregate struct {
	TotalBytes uint64      `json:"total_bytes"`
	UsedBytes  uint64      `json:"used_bytes"`
	UsedPct    float64     `json:"used_pct"`
	Devices    []DiskStats `json:"devices"`
}

type DiskStats struct {
	Mountpoint string  `json:"mountpoint"`
	Fstype     string  `json:"fstype"`
	TotalBytes uint64  `json:"total_bytes"`
	UsedBytes  uint64  `json:"used_bytes"`
	UsedPct    float64 `json:"used_pct"`
}

type NetAggregate struct {
	RxBytesPerSec float64    `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64    `json:"tx_bytes_per_sec"`
	Interfaces    []NetStats `json:"interfaces"`
}

type NetStats struct {
	Name          string  `json:"name"`
	RxBytesPerSec float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec"`
}

type Trends struct {
	CPUUsagePct []Point `json:"cpu_usage_pct"`
	MemoryPct   []Point `json:"memory_pct"`
	NetworkRx   []Point `json:"network_rx"`
	NetworkTx   []Point `json:"network_tx"`
}

type ResourceInfo struct {
	Hostname   string   `json:"hostname"`
	OS         string   `json:"os"`
	Arch       string   `json:"arch"`
	CPUCores   int      `json:"cpu_cores"`
	Disks      []string `json:"disks"`
	Interfaces []string `json:"interfaces"`
}

type Collector struct {
	mu            sync.RWMutex
	summary       Summary
	resources     ResourceInfo
	interval      time.Duration
	historyMax    int
	prevNetByName map[string]net.IOCountersStat
	prevAt        time.Time
}

func NewCollector(interval time.Duration) (*Collector, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	c := &Collector{
		interval:      interval,
		historyMax:    120,
		prevNetByName: map[string]net.IOCountersStat{},
		resources: ResourceInfo{
			Hostname: hostname,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
		},
	}

	if err := c.sample(); err != nil {
		return nil, err
	}
	c.refreshResourceCatalog()
	return c, nil
}

func (c *Collector) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.sample()
		}
	}
}

func (c *Collector) Summary() Summary {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.summary
}

func (c *Collector) Resources() ResourceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.resources
}

func (c *Collector) sample() error {
	now := time.Now().UTC()

	cpuPct, _ := cpu.Percent(0, false)
	vm, _ := mem.VirtualMemory()
	parts, _ := disk.Partitions(false)
	ios, _ := net.IOCounters(true)
	h, _ := host.Info()

	disks := make([]DiskStats, 0, len(parts))
	for _, p := range parts {
		if !isUsefulPartition(p) {
			continue
		}
		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, DiskStats{
			Mountpoint: p.Mountpoint,
			Fstype:     p.Fstype,
			TotalBytes: u.Total,
			UsedBytes:  u.Used,
			UsedPct:    u.UsedPercent,
		})
	}
	sort.Slice(disks, func(i, j int) bool { return disks[i].Mountpoint < disks[j].Mountpoint })
	diskTotal, diskUsed := chooseAggregateDisk(disks)

	dt := now.Sub(c.prevAt).Seconds()
	if c.prevAt.IsZero() || dt <= 0 {
		dt = c.interval.Seconds()
	}

	netStats := make([]NetStats, 0, len(ios))
	var aggRx, aggTx float64
	for _, io := range ios {
		if !isUsefulInterface(io.Name) {
			continue
		}
		prev := c.prevNetByName[io.Name]
		rx := float64(io.BytesRecv-prev.BytesRecv) / dt
		tx := float64(io.BytesSent-prev.BytesSent) / dt
		if c.prevAt.IsZero() {
			rx, tx = 0, 0
		}
		netStats = append(netStats, NetStats{Name: io.Name, RxBytesPerSec: rx, TxBytesPerSec: tx})
		aggRx += rx
		aggTx += tx
	}
	sort.Slice(netStats, func(i, j int) bool { return netStats[i].Name < netStats[j].Name })

	prevMap := make(map[string]net.IOCountersStat, len(ios))
	for _, io := range ios {
		prevMap[io.Name] = io
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	cpuUsage := 0.0
	if len(cpuPct) > 0 {
		cpuUsage = cpuPct[0]
	}
	memTotal := uint64(0)
	memUsed := uint64(0)
	memUsedPct := 0.0
	if vm != nil {
		memTotal = vm.Total
		memUsed = vm.Used
		memUsedPct = vm.UsedPercent
	}
	uptime := uint64(0)
	if h != nil {
		uptime = h.Uptime
	}

	diskPct := 0.0
	if diskTotal > 0 {
		diskPct = (float64(diskUsed) / float64(diskTotal)) * 100
	}

	c.summary = Summary{
		Timestamp: now,
		Hostname:  c.resources.Hostname,
		OS:        c.resources.OS,
		Arch:      c.resources.Arch,
		UptimeSec: uptime,
		CPU:       CPUStats{UsagePct: cpuUsage},
		Memory: MemoryStats{
			TotalBytes: memTotal,
			UsedBytes:  memUsed,
			UsedPct:    memUsedPct,
		},
		Disk: DiskAggregate{
			TotalBytes: diskTotal,
			UsedBytes:  diskUsed,
			UsedPct:    diskPct,
			Devices:    disks,
		},
		Network: NetAggregate{
			RxBytesPerSec: aggRx,
			TxBytesPerSec: aggTx,
			Interfaces:    netStats,
		},
		Trends: Trends{
			CPUUsagePct: appendPoint(c.summary.Trends.CPUUsagePct, Point{TS: now, Val: cpuUsage}, c.historyMax),
			MemoryPct:   appendPoint(c.summary.Trends.MemoryPct, Point{TS: now, Val: memUsedPct}, c.historyMax),
			NetworkRx:   appendPoint(c.summary.Trends.NetworkRx, Point{TS: now, Val: aggRx}, c.historyMax),
			NetworkTx:   appendPoint(c.summary.Trends.NetworkTx, Point{TS: now, Val: aggTx}, c.historyMax),
		},
	}
	c.prevNetByName = prevMap
	c.prevAt = now
	c.resources.CPUCores = runtime.NumCPU()

	if len(disks) > 0 {
		names := make([]string, 0, len(disks))
		for _, d := range disks {
			names = append(names, fmt.Sprintf("%s (%s)", d.Mountpoint, d.Fstype))
		}
		c.resources.Disks = names
	}
	if len(netStats) > 0 {
		names := make([]string, 0, len(netStats))
		for _, n := range netStats {
			names = append(names, n.Name)
		}
		c.resources.Interfaces = names
	}

	return nil
}

func (c *Collector) refreshResourceCatalog() {
	parts, _ := disk.Partitions(false)
	disks := make([]string, 0, len(parts))
	for _, p := range parts {
		if !isUsefulPartition(p) {
			continue
		}
		disks = append(disks, p.Mountpoint)
	}
	sort.Strings(disks)

	ios, _ := net.Interfaces()
	nics := make([]string, 0, len(ios))
	for _, n := range ios {
		if !isUsefulInterface(n.Name) {
			continue
		}
		nics = append(nics, n.Name)
	}
	sort.Strings(nics)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.resources.Disks = disks
	c.resources.Interfaces = nics
	c.resources.CPUCores = runtime.NumCPU()
}

func appendPoint(arr []Point, p Point, max int) []Point {
	arr = append(arr, p)
	if len(arr) > max {
		arr = arr[len(arr)-max:]
	}
	return arr
}

func chooseAggregateDisk(disks []DiskStats) (total uint64, used uint64) {
	if len(disks) == 0 {
		return 0, 0
	}

	// Prefer the root filesystem as the host-level "disk" value.
	for _, d := range disks {
		if d.Mountpoint == "/" {
			return d.TotalBytes, d.UsedBytes
		}
	}

	// Fallback: pick the largest mount to avoid APFS/container double-counting.
	best := disks[0]
	for _, d := range disks[1:] {
		if d.TotalBytes > best.TotalBytes {
			best = d
		}
	}
	return best.TotalBytes, best.UsedBytes
}

func isUsefulPartition(p disk.PartitionStat) bool {
	mp := strings.ToLower(p.Mountpoint)
	fs := strings.ToLower(p.Fstype)

	if mp == "" {
		return false
	}
	if strings.HasPrefix(mp, "/system/volumes") || strings.HasPrefix(mp, "/private/var/vm") || strings.HasPrefix(mp, "/dev") {
		return false
	}
	if strings.HasPrefix(fs, "autofs") || strings.HasPrefix(fs, "devfs") {
		return false
	}
	return true
}

func isUsefulInterface(name string) bool {
	n := strings.ToLower(name)
	if n == "" || n == "lo" || strings.HasPrefix(n, "lo") {
		return false
	}
	if strings.HasPrefix(n, "utun") || strings.HasPrefix(n, "llw") || strings.HasPrefix(n, "awdl") || strings.HasPrefix(n, "bridge") || strings.HasPrefix(n, "ap") {
		return false
	}
	return true
}
