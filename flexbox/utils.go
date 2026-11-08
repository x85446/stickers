package flexbox

import (
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// calculateRatioWithMinimum splits distribute by the ratios in matrix like calculateRatio,
// then gives every entry whose share falls under its minimum that minimum and splits what
// is left among the others, one entry at a time.
func calculateRatioWithMinimum(distribute int, matrix []int, minimumMatrix []int) (ratioDistribution []int) {
	for range matrix {
		ratioDistribution = append(ratioDistribution, 0)
	}

	dist := calculateRatio(distribute, matrix)
	for i, d := range dist {
		if minimumMatrix[i] > d {
			ratioDistribution[i] = minimumMatrix[i]

			distribute -= minimumMatrix[i]
			matrix[i] = 0
			minimumMatrix[i] = 0
			_dist := calculateRatioWithMinimum(distribute, matrix, minimumMatrix)
			for ii, _d := range _dist {
				if ii != i {
					ratioDistribution[ii] = _d
				}
			}
			break
		} else {
			ratioDistribution[i] = d
		}
	}
	return ratioDistribution
}

// calculateRatio splits distribute by the ratios in matrix, rounding each share down and
// handing the leftover one unit at a time to the entries with the largest ratios.
func calculateRatio(distribute int, matrix []int) (ratioDistribution []int) {
	if distribute == 0 {
		for range matrix {
			ratioDistribution = append(ratioDistribution, 0)
		}
		return ratioDistribution
	}

	var combinedRatios int
	for _, value := range matrix {
		combinedRatios += value
	}

	if combinedRatios > 0 {
		var remainder int
		ratioDistribution, remainder = distributeToMatrix(distribute, combinedRatios, matrix)
		if remainder > 0 {
			for index, remainderAdded := range distributeRemainder(remainder, matrix) {
				ratioDistribution[index] += remainderAdded
			}
		}
	}

	return ratioDistribution
}

// distributeToMatrix gives each entry floor(ratio/combinedRatio * distribute) and returns
// what is left. With combinedRatio set to the largest ratio instead of the sum, the
// largest entry gets the whole amount and the others a fraction of it, which is how cells
// share a row's height.
func distributeToMatrix(distribute int, combinedRatio int, matrix []int) (distribution []int, remainder int) {
	remainder = distribute
	for _, max := range matrix {
		ratioDistributionValue := int(math.Floor((float64(max) / float64(combinedRatio)) * float64(distribute)))
		distribution = append(distribution, ratioDistributionValue)
		remainder -= ratioDistributionValue

	}
	return distribution, remainder
}

// calculateMatrixRatio splits distribute between the rows of matrix using each row's
// largest ratio, which is how rows share a box's height and columns a box's width.
func calculateMatrixRatio(distribute int, matrix [][]int) (ratioDistribution []int) {
	// get matrix max ratio for each int in matrix slice
	var maxRatio []int
	for matrixIndex, ratios := range matrix {
		maxRatio = append(maxRatio, 0)
		for _, ratio := range ratios {
			if ratio > maxRatio[matrixIndex] {
				maxRatio[matrixIndex] = ratio
			}
		}
	}

	return calculateRatio(distribute, maxRatio)
}

// distributeRemainder is simple remainder distributor, it will distribute add 1 to next highest
// matrix value till it runs out of remainder to distribute, this might be improved for some more
// complex cases
func distributeRemainder(remainder int, matrixMaxRatio []int) (remainderDistribution []int) {
	for range matrixMaxRatio {
		remainderDistribution = append(remainderDistribution, 0)
	}

	distributed := 0
	for remainder > 0 {
		maxIndex := 0
		maxRatio := 0
		for index, ratio := range matrixMaxRatio {
			// skip if already expanded
			if remainderDistribution[index] > 0 {
				continue
			}
			if ratio > maxRatio {
				maxRatio = ratio
				maxIndex = index
			}
		}
		remainderDistribution[maxIndex] += 1
		distributed += 1
		remainder -= 1
	}

	return remainderDistribution
}

// clipLines keeps at most n lines of s.
func clipLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n")
}

// fitHeight trims a rendered block to height lines. It cuts from the inside, between the
// top frame (margin and border) and the bottom frame, so both survive whenever height has
// room for them; below that it keeps the top lines.
func fitHeight(block string, height, top, bottom int) string {
	lines := strings.Split(block, "\n")
	if len(lines) <= height {
		return block
	}
	if height <= 0 {
		return ""
	}
	inner := height - top - bottom
	if inner < 0 {
		return strings.Join(lines[:height], "\n")
	}
	kept := make([]string, 0, height)
	kept = append(kept, lines[:top+inner]...)
	kept = append(kept, lines[len(lines)-bottom:]...)
	return strings.Join(kept, "\n")
}

// clipWidth truncates every line of s to width columns, keeping escape sequences intact.
func clipWidth(s string, width int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, max(width, 0), "")
		}
	}
	return strings.Join(lines, "\n")
}
