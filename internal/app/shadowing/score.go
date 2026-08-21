package shadowing

import (
	"regexp"
	"strings"
)

var punctuationRe = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)

// normalizeWords lowercases, strips punctuation and splits text into words
// so the reference sentence and the transcribed audio can be compared
// word-by-word regardless of casing/punctuation differences.
func normalizeWords(text string) []string {
	cleaned := punctuationRe.ReplaceAllString(strings.ToLower(text), "")
	return strings.Fields(cleaned)
}

// scoreTranscript compares the transcribed audio against the reference
// sentence using a word-level LCS (longest common subsequence), so word
// order matters but small insertions/omissions elsewhere don't tank the
// whole score. It returns a 0-100 score plus which reference words were
// matched vs. missed.
func scoreTranscript(reference, transcript string) (score int, correctWords, mispronouncedWords []string) {
	refWords := normalizeWords(reference)
	gotWords := normalizeWords(transcript)

	if len(refWords) == 0 {
		return 0, nil, nil
	}

	matched := wordLCSMatch(refWords, gotWords)

	correctWords = make([]string, 0, len(refWords))
	mispronouncedWords = make([]string, 0, len(refWords))
	for i, w := range refWords {
		if matched[i] {
			correctWords = append(correctWords, w)
		} else {
			mispronouncedWords = append(mispronouncedWords, w)
		}
	}

	score = (len(correctWords) * 100) / len(refWords)
	return score, correctWords, mispronouncedWords
}

// wordLCSMatch returns, for each index in a, whether that word participates
// in the longest common subsequence between a and b.
func wordLCSMatch(a, b []string) []bool {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	matched := make([]bool, n)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			matched[i] = true
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			i++
		default:
			j++
		}
	}
	return matched
}
