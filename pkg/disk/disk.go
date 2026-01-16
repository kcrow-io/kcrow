package disk

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getOverlayPath(containerRootfs string) (uint64, string, error) {
	f, err := os.Open(SystemMountInfoFile)
	if err != nil {
		return 0, "", fmt.Errorf("failed to open host_mountinfo: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, containerRootfs) {
			continue
		}

		fields := strings.Split(line, " - ")
		if len(fields) < 2 {
			continue
		}

		preFields := strings.Fields(fields[0])
		mountPoint := preFields[4]

		if mountPoint != containerRootfs {
			continue
		}

		postFields := strings.Fields(fields[1])
		if len(postFields) < 3 {
			continue
		}
		options := postFields[2]
		for _, opt := range strings.Split(options, ",") {
			if strings.HasPrefix(opt, "upperdir=") {
				upperDir := strings.TrimPrefix(opt, "upperdir=")
				cleanPath := filepath.Clean(upperDir)
				if strings.HasSuffix(cleanPath, "/fs") {
					cleanPath = filepath.Dir(cleanPath)
				}
				snapshotId, err := strconv.ParseUint(filepath.Base(cleanPath), 10, 64)
				if err != nil {
					return 0, "", fmt.Errorf("failed to parse snapshot id from path [%s]: %v", upperDir, err)
				}

				return snapshotId, upperDir, nil
			}
		}
	}

	return 0, "", fmt.Errorf("overlay path not found in mountinfo for %s", containerRootfs)
}

func applyXFSQuota(projectID uint64, path string, limitMB int) error {
	mountPoint := ContainerdRootPath
	exec.Command("xfs_quota", "-x", "-c", fmt.Sprintf("project -C -p %s %d", path, projectID), mountPoint).Run()

	setupCmd := exec.Command("xfs_quota", "-x", "-c", fmt.Sprintf("project -s -p %s %d", path, projectID), mountPoint)
	if out, err := setupCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("setup project failed: %s, %v", string(out), err)
	}

	workpath := getWorkPath(path)

	workCmd := exec.Command("xfs_quota", "-x", "-c", fmt.Sprintf("project -s -p %s %d", workpath, projectID), mountPoint)
	if out, err := workCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("setup project failed: %s, %v", string(out), err)
	}

	limitCmd := exec.Command("xfs_quota", "-x", "-c", fmt.Sprintf("limit -p bhard=%dm %d", limitMB, projectID), mountPoint)
	if out, err := limitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("set limit failed: %s, %v", string(out), err)
	}
	return nil
}

func getWorkPath(fsPath string) string {
	return strings.TrimSuffix(fsPath, "/fs") + "/work"
}

func checkContainerdRootPathQuotaEnabled() bool {
	data, _ := os.ReadFile("/proc/mounts")
	mountPoint := "/var/lib/containerd"

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[1] == mountPoint {
			opts := "," + fields[3] + ","
			// Due to the current support of only the xfs file system, only Prjquota is determined, and the ext4 file system is identified as pquota
			return strings.Contains(opts, ",prjquota,")
		}
	}
	return false
}
