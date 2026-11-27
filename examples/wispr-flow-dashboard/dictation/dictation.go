// Package dictation is the dictation page: the welcome header, the style
// hero, the stats, and the history list.
package dictation

// Data is the data the dictation page prints.
type Data struct {
	User       string
	TotalWords string
	WPM        string
	Streak     string
	Rows       []DictRow
}

// DictRow is one dictation-history row.
type DictRow struct {
	Time string
	Text string
}

// Default returns the dictation page data. TotalWords references the local
// Wispr usage aggregate; the history rows remain sample content.
func Default() Data {
	return Data{
		User:       "Chinmay",
		TotalWords: "424.6K",
		WPM:        "145",
		Streak:     "53",
		Rows: []DictRow{
			{"9:18pm", "We have some commits. Are those related to us? If not, even " +
				"let's commit and push them."},
			{"9:16pm", "Make the comment detailed."},
			{"9:16pm", "Please commit and push the current changes which we have."},
			{"9:15pm", "The minimum width that you have added seems to be not working " +
				"because whenever I go below 300 it is just squishing the elements. " +
				"It doesn't look good."},
			{"9:10pm", "earlier, we were discussing the benchmarks and stuff, right? " +
				"Okay, talk about"},
			{"9:09pm", "Can you please modify the developer tools such that they " +
				"should include the CPU consumed."},
		},
	}
}
