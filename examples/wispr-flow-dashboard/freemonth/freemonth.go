// Package freemonth is the get-a-free-month page: referral progress and the reward.
package freemonth

// Data is the data the free-month page prints.
type Data struct {
	Joined   int
	Goal     int
	BarWidth string
	Hint     string
	Reward   string
	Unlocked bool
	Steps    []FreeStep
}

// FreeStep is one row in the free-month progress list.
type FreeStep struct {
	ID    string
	Label string
	Done  bool
}

// Default returns the free-month page data.
func Default() Data {
	return Data{
		Joined:   defaultJoined,
		Goal:     defaultGoal,
		BarWidth: "66%",
		Hint:     "1 more to go",
		Reward:   "1 month of Flow Pro",
		Unlocked: false,
		Steps: []FreeStep{
			{"free-step-invite", "Invite sent", true},
			{"free-step-joined", "Friend joined", true},
			{"free-step-reward", "Reward unlocked", false},
		},
	}
}
