package player

import "strconv"

// coverJob names one artwork download.
type coverJob struct {
	name string
	url  string
}

// maxCoverWorkers caps how many artwork downloads run at once.
const maxCoverWorkers = 12

// coverJobs maps the six songs and twelve albums onto the fixed slots.
func coverJobs(songs, albums []searchResult) []coverJob {
	jobs := make([]coverJob, 0, len(songs)+len(albums))

	for i, s := range songs {
		jobs = append(jobs, coverJob{name: "track-" + strconv.Itoa(i), url: upscale(s.Artwork)})
	}

	for i, al := range albums {
		name := "pick-" + strconv.Itoa(i)
		if i >= coverSplit {
			name = "card-" + strconv.Itoa(i-coverSplit)
		}

		jobs = append(jobs, coverJob{name: name, url: upscale(al.Artwork)})
	}

	return jobs
}
