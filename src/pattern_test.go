package fzf

import (
	"reflect"
	"testing"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

var slab *util.Slab

func init() {
	slab = util.MakeSlab(slab16Size, slab32Size)
}

func TestParseTermsExtended(t *testing.T) {
	terms := parseTerms(true, CaseSmart, false,
		"aaa 'bbb ^ccc ddd$ !eee !'fff !^ggg !hhh$ | ^iii$ ^xxx | 'yyy | zzz$ | !ZZZ |")
	if len(terms) != 9 ||
		terms[0][0].typ != termFuzzy || terms[0][0].inv ||
		terms[1][0].typ != termExact || terms[1][0].inv ||
		terms[2][0].typ != termPrefix || terms[2][0].inv ||
		terms[3][0].typ != termSuffix || terms[3][0].inv ||
		terms[4][0].typ != termExact || !terms[4][0].inv ||
		terms[5][0].typ != termFuzzy || !terms[5][0].inv ||
		terms[6][0].typ != termPrefix || !terms[6][0].inv ||
		terms[7][0].typ != termSuffix || !terms[7][0].inv ||
		terms[7][1].typ != termEqual || terms[7][1].inv ||
		terms[8][0].typ != termPrefix || terms[8][0].inv ||
		terms[8][1].typ != termExact || terms[8][1].inv ||
		terms[8][2].typ != termSuffix || terms[8][2].inv ||
		terms[8][3].typ != termExact || !terms[8][3].inv {
		t.Errorf("%v", terms)
	}
	for _, termSet := range terms[:8] {
		term := termSet[0]
		if len(term.text) != 3 {
			t.Errorf("%v", term)
		}
	}
}

func TestParseTermsExtendedExact(t *testing.T) {
	terms := parseTerms(false, CaseSmart, false,
		"aaa 'bbb ^ccc ddd$ !eee !'fff !^ggg !hhh$")
	if len(terms) != 8 ||
		terms[0][0].typ != termExact || terms[0][0].inv || len(terms[0][0].text) != 3 ||
		terms[1][0].typ != termFuzzy || terms[1][0].inv || len(terms[1][0].text) != 3 ||
		terms[2][0].typ != termPrefix || terms[2][0].inv || len(terms[2][0].text) != 3 ||
		terms[3][0].typ != termSuffix || terms[3][0].inv || len(terms[3][0].text) != 3 ||
		terms[4][0].typ != termExact || !terms[4][0].inv || len(terms[4][0].text) != 3 ||
		terms[5][0].typ != termFuzzy || !terms[5][0].inv || len(terms[5][0].text) != 3 ||
		terms[6][0].typ != termPrefix || !terms[6][0].inv || len(terms[6][0].text) != 3 ||
		terms[7][0].typ != termSuffix || !terms[7][0].inv || len(terms[7][0].text) != 3 {
		t.Errorf("%v", terms)
	}
}

func TestParseTermsEmpty(t *testing.T) {
	terms := parseTerms(true, CaseSmart, false, "' ^ !' !^")
	if len(terms) != 0 {
		t.Errorf("%v", terms)
	}
}

func buildPattern(fuzzy bool, fuzzyAlgo algo.Algo, extended bool, caseMode Case, normalize bool, forward bool,
	withPos bool, cacheable bool, nth []Range, delimiter Delimiter, runes []rune, keymapConvert bool) *Pattern {
	return BuildPattern(NewChunkCache(), make(map[string]*Pattern),
		fuzzy, fuzzyAlgo, extended, caseMode, normalize, forward,
		withPos, cacheable, nth, delimiter, revision{}, runes, nil, keymapConvert)
}

func TestExact(t *testing.T) {
	pattern := buildPattern(true, algo.FuzzyMatchV2, true, CaseSmart, false, true,
		false, true, []Range{}, Delimiter{}, []rune("'abc"), false)
	chars := util.ToChars([]byte("aabbcc abc"))
	res, pos := algo.ExactMatchNaive(
		pattern.caseSensitive, pattern.normalize, pattern.forward, &chars, pattern.termSets[0][0].text, true, nil)
	if res.Start != 7 || res.End != 10 {
		t.Errorf("%v / %d / %d", pattern.termSets, res.Start, res.End)
	}
	if pos != nil {
		t.Errorf("pos is expected to be nil")
	}
}

func TestEqual(t *testing.T) {
	pattern := buildPattern(true, algo.FuzzyMatchV2, true, CaseSmart, false, true, false, true, []Range{}, Delimiter{}, []rune("^AbC$"), false)

	match := func(str string, sidxExpected int, eidxExpected int) {
		chars := util.ToChars([]byte(str))
		res, pos := algo.EqualMatch(
			pattern.caseSensitive, pattern.normalize, pattern.forward, &chars, pattern.termSets[0][0].text, true, nil)
		if res.Start != sidxExpected || res.End != eidxExpected {
			t.Errorf("%v / %d / %d", pattern.termSets, res.Start, res.End)
		}
		if pos != nil {
			t.Errorf("pos is expected to be nil")
		}
	}
	match("ABC", -1, -1)
	match("AbC", 0, 3)
	match("AbC  ", 0, 3)
	match(" AbC ", 1, 4)
	match("  AbC", 2, 5)
}

func TestCaseSensitivity(t *testing.T) {
	pat1 := buildPattern(true, algo.FuzzyMatchV2, false, CaseSmart, false, true, false, true, []Range{}, Delimiter{}, []rune("abc"), false)
	pat2 := buildPattern(true, algo.FuzzyMatchV2, false, CaseSmart, false, true, false, true, []Range{}, Delimiter{}, []rune("Abc"), false)
	pat3 := buildPattern(true, algo.FuzzyMatchV2, false, CaseIgnore, false, true, false, true, []Range{}, Delimiter{}, []rune("abc"), false)
	pat4 := buildPattern(true, algo.FuzzyMatchV2, false, CaseIgnore, false, true, false, true, []Range{}, Delimiter{}, []rune("Abc"), false)
	pat5 := buildPattern(true, algo.FuzzyMatchV2, false, CaseRespect, false, true, false, true, []Range{}, Delimiter{}, []rune("abc"), false)
	pat6 := buildPattern(true, algo.FuzzyMatchV2, false, CaseRespect, false, true, false, true, []Range{}, Delimiter{}, []rune("Abc"), false)

	if string(pat1.text) != "abc" || pat1.caseSensitive != false ||
		string(pat2.text) != "Abc" || pat2.caseSensitive != true ||
		string(pat3.text) != "abc" || pat3.caseSensitive != false ||
		string(pat4.text) != "abc" || pat4.caseSensitive != false ||
		string(pat5.text) != "abc" || pat5.caseSensitive != true ||
		string(pat6.text) != "Abc" || pat6.caseSensitive != true {
		t.Error("Invalid case conversion")
	}
}

func TestOrigTextAndTransformed(t *testing.T) {
	pattern := buildPattern(true, algo.FuzzyMatchV2, true, CaseSmart, false, true, false, true, []Range{}, Delimiter{}, []rune("jg"), false)
	tokens := Tokenize("junegunn", Delimiter{})
	trans := Transform(tokens, []Range{{1, 1}})

	origBytes := []byte("junegunn.choi")
	for _, extended := range []bool{false, true} {
		chunk := Chunk{count: 1}
		chunk.items[0] = Item{
			text:        util.ToChars([]byte("junegunn")),
			origText:    &origBytes,
			transformed: &transformed{pattern.revision, trans}}
		pattern.extended = extended
		matches := pattern.matchChunk(&chunk, nil, slab) // No cache
		if !(matches[0].item.text.ToString() == "junegunn" &&
			string(*matches[0].item.origText) == "junegunn.choi" &&
			reflect.DeepEqual((*matches[0].item.transformed).tokens, trans)) {
			t.Error("Invalid match result", matches)
		}

		match, offsets, pos := pattern.MatchItem(&chunk.items[0], true, slab)
		if !(match.item.text.ToString() == "junegunn" &&
			string(*match.item.origText) == "junegunn.choi" &&
			offsets[0][0] == 0 && offsets[0][1] == 5 &&
			reflect.DeepEqual((*match.item.transformed).tokens, trans)) {
			t.Error("Invalid match result", match, offsets, extended)
		}
		if !((*pos)[0] == 4 && (*pos)[1] == 0) {
			t.Error("Invalid pos array", *pos)
		}
	}
}

func TestCacheKey(t *testing.T) {
	test := func(extended bool, patStr string, expected string, cacheable bool) {
		pat := buildPattern(true, algo.FuzzyMatchV2, extended, CaseSmart, false, true, false, true, []Range{}, Delimiter{}, []rune(patStr), false)
		if pat.CacheKey() != expected {
			t.Errorf("Expected: %s, actual: %s", expected, pat.CacheKey())
		}
		if pat.cacheable != cacheable {
			t.Errorf("Expected: %t, actual: %t (%s)", cacheable, pat.cacheable, patStr)
		}
	}
	test(false, "foo !bar", "foo !bar", true)
	test(false, "foo | bar !baz", "foo | bar !baz", true)
	test(true, "foo  bar  baz", "foo\tbar\tbaz", true)
	test(true, "foo !bar", "foo", false)
	test(true, "foo !bar   baz", "foo\tbaz", false)
	test(true, "foo | bar baz", "baz", false)
	test(true, "foo | bar | baz", "", false)
	test(true, "foo | bar !baz", "", false)
	test(true, "| | foo", "", false)
	test(true, "| | | foo", "foo", false)
}

func TestCacheable(t *testing.T) {
	test := func(fuzzy bool, str string, expected string, cacheable bool) {
		pat := buildPattern(fuzzy, algo.FuzzyMatchV2, true, CaseSmart, true, true, false, true, []Range{}, Delimiter{}, []rune(str), false)
		if pat.CacheKey() != expected {
			t.Errorf("Expected: %s, actual: %s", expected, pat.CacheKey())
		}
		if cacheable != pat.cacheable {
			t.Errorf("Invalid Pattern.cacheable for \"%s\": %v (expected: %v)", str, pat.cacheable, cacheable)
		}
	}
	test(true, "foo bar", "foo\tbar", true)
	test(true, "foo 'bar", "foo\tbar", false)
	test(true, "foo !bar", "foo", false)

	test(false, "foo bar", "foo\tbar", true)
	test(false, "foo 'bar", "foo", false)
	test(false, "foo '", "foo", true)
	test(false, "foo 'bar", "foo", false)
	test(false, "foo !bar", "foo", false)
}
