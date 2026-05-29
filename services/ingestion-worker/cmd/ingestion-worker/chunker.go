package main

import "strings"

// chunkText splits text into overlapping chunks (by character count).
// It prefers splitting on paragraph boundaries (blank lines) but will fall back
// to slicing long paragraphs.
func chunkText(text string, chunkSize int, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		return []string{text}
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4
	}

	paragraphs := splitParagraphs(text)
	var out []string
	var current strings.Builder

	flush := func() {
		s := strings.TrimSpace(current.String())
		if s != "" {
			out = append(out, s)
		}
		current.Reset()
	}

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// If paragraph itself is huge, split it directly.
		if runeLen(p) > chunkSize {
			flush()
			out = append(out, chunkByWindow(p, chunkSize, overlap)...)
			continue
		}

		candidate := p
		if current.Len() > 0 {
			candidate = current.String() + "\n\n" + p
		}

		if runeLen(candidate) <= chunkSize {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(p)
			continue
		}

		flush()
		current.WriteString(p)
	}
	flush()
	return out
}

func splitParagraphs(text string) []string {
	// Normalize newlines then split on blank lines.
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var paras []string
	var buf strings.Builder
	lines := strings.Split(text, "\n")

	flush := func() {
		p := strings.TrimSpace(buf.String())
		if p != "" {
			paras = append(paras, p)
		}
		buf.Reset()
	}

	emptyStreak := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyStreak++
			if emptyStreak >= 1 {
				flush()
			}
			continue
		}
		emptyStreak = 0
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(line)
	}
	flush()
	return paras
}

func chunkByWindow(text string, size int, overlap int) []string {
	r := []rune(text)
	if len(r) <= size {
		return []string{strings.TrimSpace(text)}
	}
	step := size - overlap
	if step <= 0 {
		step = size
	}

	var out []string
	for start := 0; start < len(r); start += step {
		end := start + size
		if end > len(r) {
			end = len(r)
		}
		chunk := strings.TrimSpace(string(r[start:end]))
		if chunk != "" {
			out = append(out, chunk)
		}
		if end == len(r) {
			break
		}
	}
	return out
}

func runeLen(s string) int {
	return len([]rune(s))
}
