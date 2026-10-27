package main

// Graph cell kinds. A row holds one cell per lane.
const (
	graphNode   = "node"
	graphVert   = "vert"
	graphBranch = "branch"
	graphMerge  = "merge"
	graphSpace  = "space"
)

// maxCommitLanes bounds the history graph width so a wide merge cannot push the
// commit list off the rail.
const maxCommitLanes = 6

// graphCell is one lane of one history row.
type graphCell struct {
	Lane int
	Kind string
}

// commitGraph is the lane layout for a newest-first commit list. Rows holds the
// cells of each commit's row keyed by full hash, and Lanes is the widest row.
// graphCells pads a row to Lanes columns before drawing.
//
// Kinds:
//   - node: the commit dot, exactly one per row.
//   - vert: a lane passing through the row toward an older commit.
//   - branch: the row opens a lane for an extra parent.
//   - merge: the row joins a lane that is already open, either an extra parent
//     another line already tracks or a second child line ending at this commit.
//   - space: no line in the lane.
type commitGraph struct {
	Lanes int
	Rows  map[string][]graphCell
}

// buildCommitGraph lays commits out in lanes. Commits arrive newest first, as
// git log prints them. The first parent keeps the commit's lane, an extra parent
// opens the next free lane, and a parent another lane already tracks is joined
// instead of duplicated. The width stops at maxCommitLanes; once every lane is
// busy, a new line shares the last lane.
func buildCommitGraph(commits []gitCommit) commitGraph {
	graph := commitGraph{Rows: make(map[string][]graphCell, len(commits))}
	if len(commits) == 0 {
		return graph
	}

	// laneParent[i] is the hash lane i draws toward next. An empty entry is a
	// free lane.
	laneParent := []string{""}

	for i := range commits {
		commit := commits[i]

		lane := laneFor(laneParent, commit.Hash)
		if lane < 0 {
			lane = reserveLane(&laneParent)
			laneParent[lane] = commit.Hash
		}

		cells := graphCells(nil, len(laneParent))

		for other := range laneParent {
			switch {
			case other == lane, laneParent[other] == "":
			case laneParent[other] == commit.Hash:
				// A child line ends here because the commit has two children.
				cells[other] = graphCell{Lane: other, Kind: graphMerge}
				laneParent[other] = ""
			default:
				cells[other] = graphCell{Lane: other, Kind: graphVert}
			}
		}

		if len(commit.Parents) > 1 {
			for _, parent := range commit.Parents[1:] {
				if other := laneFor(laneParent, parent); other >= 0 {
					cells[other] = graphCell{Lane: other, Kind: graphMerge}

					continue
				}

				other := reserveLane(&laneParent)
				laneParent[other] = parent
				cells = graphCells(cells, len(laneParent))
				cells[other] = graphCell{Lane: other, Kind: graphBranch}
			}
		}

		cells[lane] = graphCell{Lane: lane, Kind: graphNode}

		if len(commit.Parents) == 0 {
			laneParent[lane] = ""
		} else {
			laneParent[lane] = commit.Parents[0]
		}

		graph.Rows[commit.Hash] = cells
	}

	graph.Lanes = len(laneParent)

	return graph
}

// laneFor returns the lane already drawing toward hash, or -1.
func laneFor(lanes []string, hash string) int {
	for i, waiting := range lanes {
		if waiting != "" && waiting == hash {
			return i
		}
	}

	return -1
}

// reserveLane returns a free lane and extends the slice when it opens a new
// column. With every lane busy it returns the last lane so the graph stays
// inside maxCommitLanes.
func reserveLane(lanes *[]string) int {
	for i, waiting := range *lanes {
		if waiting == "" {
			return i
		}
	}

	if len(*lanes) < maxCommitLanes {
		*lanes = append(*lanes, "")
	}

	return len(*lanes) - 1
}

// graphCells pads row to lanes columns. Missing cells become space and every
// returned cell carries its column as Lane, so a renderer can walk the result by
// index.
func graphCells(row []graphCell, lanes int) []graphCell {
	cells := make([]graphCell, lanes)
	for i := range cells {
		cells[i] = graphCell{Lane: i, Kind: graphSpace}
	}

	for _, cell := range row {
		if cell.Lane < 0 || cell.Lane >= lanes {
			continue
		}

		cells[cell.Lane] = graphCell{Lane: cell.Lane, Kind: cell.Kind}
	}

	return cells
}
