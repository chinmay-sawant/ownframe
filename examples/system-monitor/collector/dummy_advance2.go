package collector

// dmDisk is one simulated disk: space plus cumulative IO.
type dmDisk struct {
	name, mount string
	total, free uint64
	read, write uint64
}

// dmNet is one simulated interface with cumulative counters.
type dmNet struct {
	name   string
	up     bool
	rx, tx uint64
	errs   uint64
}

// advanceDisks grows IO counters and walks free space. Every diskBurstEvery
// tick writes a burst, so parsing and graphs see a spike.
func (d *Dummy) advanceDisks() {
	for i := range d.disks {
		write := d.rng.Int63n(diskWriteSpan)
		if d.tick%diskBurstEvery == 0 {
			write += diskBurst
		}
		d.disks[i].read += uint64(d.rng.Int63n(diskReadSpan)) //nolint:gosec // reason: Int63n never returns a negative value.
		d.disks[i].write += uint64(write)                     //nolint:gosec // reason: the counter only grows from zero, so it stays non-negative.

		step := int64(d.rng.Intn(diskWobble)) - diskWobbleBase
		d.disks[i].free = clampU64(int64(d.disks[i].free)+step, diskFreeFloor, int64(d.disks[i].total)) //nolint:gosec // reason: disk space stays within tens of gigabytes, far below MaxInt64.
	}
}

// advanceNets grows interface counters, adds a USB interface once, and raises
// an occasional error.
func (d *Dummy) advanceNets() {
	if d.tick == usbAtTick {
		d.nets = append(d.nets, dmNet{name: "usb0", up: true})
	}

	for i := range d.nets {
		rx := d.rng.Int63n(netSpan)
		if d.tick%netBurstEvery == 0 {
			rx += netBurst
		}
		d.nets[i].rx += uint64(rx)                      //nolint:gosec // reason: the counter only grows from zero, so it stays non-negative.
		d.nets[i].tx += uint64(d.rng.Int63n(netTxSpan)) //nolint:gosec // reason: Int63n never returns a negative value.

		if d.rng.Float64() < netErrRate {
			d.nets[i].errs++
		}
	}
}
