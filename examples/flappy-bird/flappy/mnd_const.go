package flappy

// Collision, bob, layout, and tick numbers for the game.
const (
	collideDX      = 15
	collideDY      = 11
	collideW       = 30
	collideH       = 22
	halfDiv        = 2
	bobFreq        = 3.2
	bobAmp         = 8
	pcgStream      = 0x9e3779b97f4a7c15
	cloudBase      = 40
	cloudSpacing   = 170
	stripeSpacing  = 80
	offscreen      = -1000
	wingRate       = 12
	cloudDrift     = 0.18
	cloudResetX    = 60
	cloudResetGap  = 40
	stripeResetX   = 30
	stripeResetGap = 12
	spawnOffset    = 30
	maxDT          = 0.1
)
