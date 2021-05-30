package subsystems_test

import (
	"testing"

	subsystems "github.com/YasinZhangX/dockerSrc/cgroups/subsystems"
)

func TestFindCgroupMountPoint(t *testing.T) {
	t.Logf("cpu subsystem mount point %v\n", subsystems.FindCgroupMountpoint("cpu"))
	t.Logf("cpuset subsystem mount point %v\n", subsystems.FindCgroupMountpoint("cpuset"))
	t.Logf("memory subsystem mount point %v\n", subsystems.FindCgroupMountpoint("memory"))
}
