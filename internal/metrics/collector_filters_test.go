package metrics

import (
	"testing"

	"github.com/shirou/gopsutil/v4/disk"
)

func TestIsUsefulPartition(t *testing.T) {
	tests := []struct {
		name string
		part disk.PartitionStat
		want bool
	}{
		{
			name: "keeps root apfs",
			part: disk.PartitionStat{Mountpoint: "/", Fstype: "apfs"},
			want: true,
		},
		{
			name: "drops system volumes",
			part: disk.PartitionStat{Mountpoint: "/System/Volumes/Data", Fstype: "apfs"},
			want: false,
		},
		{
			name: "drops devfs",
			part: disk.PartitionStat{Mountpoint: "/dev", Fstype: "devfs"},
			want: false,
		},
		{
			name: "drops autofs mounts",
			part: disk.PartitionStat{Mountpoint: "/net", Fstype: "autofs"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUsefulPartition(tt.part)
			if got != tt.want {
				t.Fatalf("isUsefulPartition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUsefulInterface(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		want  bool
	}{
		{name: "keeps en0", iface: "en0", want: true},
		{name: "drops loopback", iface: "lo0", want: false},
		{name: "drops utun", iface: "utun2", want: false},
		{name: "drops awdl", iface: "awdl0", want: false},
		{name: "drops bridge", iface: "bridge0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUsefulInterface(tt.iface)
			if got != tt.want {
				t.Fatalf("isUsefulInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChooseAggregateDiskPrefersRootMount(t *testing.T) {
	disks := []DiskStats{
		{Mountpoint: "/Volumes/External", TotalBytes: 2_000, UsedBytes: 700},
		{Mountpoint: "/", TotalBytes: 500, UsedBytes: 300},
	}

	total, used := chooseAggregateDisk(disks)
	if total != 500 || used != 300 {
		t.Fatalf("expected root disk totals, got total=%d used=%d", total, used)
	}
}

func TestChooseAggregateDiskFallsBackToLargest(t *testing.T) {
	disks := []DiskStats{
		{Mountpoint: "/mnt/a", TotalBytes: 100, UsedBytes: 40},
		{Mountpoint: "/mnt/b", TotalBytes: 800, UsedBytes: 120},
		{Mountpoint: "/mnt/c", TotalBytes: 400, UsedBytes: 90},
	}

	total, used := chooseAggregateDisk(disks)
	if total != 800 || used != 120 {
		t.Fatalf("expected largest disk totals, got total=%d used=%d", total, used)
	}
}
