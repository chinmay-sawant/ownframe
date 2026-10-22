//go:build linux

package collector

import "syscall"

// statfsT holds the fields space accounting needs from statfs.
type statfsT struct {
	bsize  int64
	blocks uint64
	bavail uint64
}

// statfsCall reads filesystem space through statfs. Free space counts blocks
// available to an unprivileged caller, which is what a user sees as free.
func statfsCall(path string, out *statfsT) error {
	var st syscall.Statfs_t

	if err := syscall.Statfs(path, &st); err != nil {
		return err
	}

	out.bsize = st.Bsize
	out.blocks = st.Blocks
	out.bavail = st.Bavail

	return nil
}

// statfs returns total and free bytes for a mount point.
func statfs(path string) (uint64, uint64, error) {
	var st statfsT

	if err := statfsCall(path, &st); err != nil {
		return 0, 0, err
	}

	block := uint64(st.bsize) //nolint:gosec // reason: filesystem block sizes are small positive values.

	return st.blocks * block, st.bavail * block, nil
}
