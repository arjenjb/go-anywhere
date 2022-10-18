package main

func Matches(search string, input string) bool {
	if len(search) == 0 {
		return true
	}

	si := 0
	ii := 0

	// Keep searching as long as there is input left AND the search input is not longer than the remaining input
	for ii < len(input) && (len(search)-si <= len(input)-ii) {
		if search[si] == input[ii] {
			si++
			// Have we matched all search characters to the input string?
			if si == len(search) {
				return true
			}
		}
		ii++
	}

	return false
}

// MatchScore returns a float indicating how well a string is matched by input
func MatchScore(search string, input string) float64 {
	si := 0
	ii := 0

	// Track the number of matched runs
	li := -2 // The last matched input index
	runs := 0

	fullMatch := false

	// Keep searching as long as there is input left AND the search input is not longer than the remaining input
	for ii < len(input) && len(search)-si > 0 && (len(search)-si <= len(input)-ii) {
		if search[si] == input[ii] {
			if li+1 != ii {
				runs++
			}
			li = ii // Track the last matched index

			si++
			// Have we matched all search characters to the input string?
			if si == len(search) {
				fullMatch = true
				break
			}
		}

		ii++
	}

	if !fullMatch {
		return 0.0
	}

	matchPercentage := float64(si) / float64(len(input))

	if runs <= 1 {
		return matchPercentage
	}

	unmatchedPercentage := 1.0 - matchPercentage
	correction := (1.0 - (1.0 / float64(runs))) * unmatchedPercentage * matchPercentage

	return matchPercentage - correction
}
