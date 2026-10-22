package cat

// SetRandomBehavior changes idle behavior without replacing the latest message.
func (c *Companion) SetRandomBehavior(enabled bool) {
	c.RandomBehavior = enabled
	c.animation.random = enabled
	if !c.shown {
		c.animation.cycle = enabled
		c.animation.step = int(c.seconds / stepInterval)
	}
}
