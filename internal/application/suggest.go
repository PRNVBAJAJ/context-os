package application

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
)

// MemorySuggestion is a candidate memory entry derived from checkpoint notes.
type MemorySuggestion struct {
	// Phrase is the suggested memory key / topic.
	Phrase string
	// Occurrences is how many checkpoint notes contained this phrase.
	Occurrences int
	// ExampleNote is one checkpoint note where the phrase appeared.
	ExampleNote string
}

// SuggestMemoriesOptions controls which checkpoints are analyzed.
type SuggestMemoriesOptions struct {
	RootPath string
	// TopN caps how many suggestions to return. 0 defaults to 10.
	TopN int
}

// SuggestMemories analyzes checkpoint notes and proposes memory entries based
// on recurring phrases. It uses simple frequency analysis — no LLM required.
func SuggestMemories(ctx context.Context, opts SuggestMemoriesOptions) ([]MemorySuggestion, error) {
	if opts.TopN == 0 {
		opts.TopN = 10
	}

	dbPath := filepath.Join(project.Dir(opts.RootPath), "runtime.db")
	store, err := storage.Open(ctx, dbPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = store.Close() }()

	p, err := store.Projects().GetByPath(ctx, opts.RootPath)
	if err != nil {
		return nil, err
	}

	// Load all checkpoints (across all workflows).
	cps, err := store.Checkpoints().List(ctx, p.ID, storage.CheckpointFilter{})
	if err != nil {
		return nil, err
	}

	// Collect non-empty, non-auto notes.
	var notes []string
	for _, cp := range cps {
		n := strings.TrimSpace(cp.Note)
		if n != "" && n != "auto" {
			notes = append(notes, n)
		}
	}

	if len(notes) == 0 {
		return nil, nil
	}

	// Load existing memory keys so we don't suggest what already exists.
	memories, err := store.Memories().List(ctx, p.ID, storage.MemoryFilter{})
	if err != nil {
		return nil, err
	}
	existing := make(map[string]bool, len(memories))
	for _, m := range memories {
		existing[m.Key] = true
		for _, w := range tokenize(m.Key) {
			existing[w] = true
		}
	}

	// Count bigram and trigram frequencies across checkpoint notes.
	type phraseEntry struct {
		count   int
		example string
	}
	freq := map[string]*phraseEntry{}

	for _, note := range notes {
		tokens := tokenize(note)
		seen := map[string]bool{}

		// Bigrams
		for i := 0; i < len(tokens)-1; i++ {
			bg := tokens[i] + "-" + tokens[i+1]
			if len(bg) >= 8 && !seen[bg] {
				seen[bg] = true
				if freq[bg] == nil {
					freq[bg] = &phraseEntry{}
				}
				freq[bg].count++
				if freq[bg].example == "" {
					freq[bg].example = note
				}
			}
		}

		// Unigrams (longer meaningful words only)
		for _, tok := range tokens {
			if len(tok) >= 7 && !seen[tok] {
				seen[tok] = true
				if freq[tok] == nil {
					freq[tok] = &phraseEntry{}
				}
				freq[tok].count++
				if freq[tok].example == "" {
					freq[tok].example = note
				}
			}
		}
	}

	// Collect candidates that appeared in at least 2 notes and aren't already in memory.
	type candidate struct {
		phrase      string
		occurrences int
		example     string
	}
	var candidates []candidate
	for phrase, e := range freq {
		if e.count < 2 {
			continue
		}
		// Skip phrases already covered by existing memories.
		if existing[phrase] {
			continue
		}
		candidates = append(candidates, candidate{phrase, e.count, e.example})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].occurrences != candidates[j].occurrences {
			return candidates[i].occurrences > candidates[j].occurrences
		}
		return candidates[i].phrase < candidates[j].phrase
	})

	if len(candidates) > opts.TopN {
		candidates = candidates[:opts.TopN]
	}

	out := make([]MemorySuggestion, len(candidates))
	for i, c := range candidates {
		out[i] = MemorySuggestion{
			Phrase:      c.phrase,
			Occurrences: c.occurrences,
			ExampleNote: c.example,
		}
	}
	return out, nil
}

// stopWords are common English words that carry no project-specific meaning.
var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "from": true,
	"that": true, "this": true, "are": true, "was": true, "have": true,
	"been": true, "not": true, "will": true, "all": true, "but": true,
	"into": true, "its": true, "now": true, "new": true, "add": true,
	"run": true, "use": true, "set": true, "can": true, "via": true,
	"also": true, "using": true, "done": true, "next": true, "test": true,
	"both": true, "each": true, "when": true, "then": true, "after": true,
	"before": true, "pass": true, "working": true, "tests": true,
	"added": true, "updated": true, "fixed": true, "impl": true,
	"context": true, "workflow": true, "memory": true, "checkpoint": true,
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var out []string
	for _, f := range fields {
		if !stopWords[f] && len(f) >= 3 {
			out = append(out, f)
		}
	}
	return out
}
