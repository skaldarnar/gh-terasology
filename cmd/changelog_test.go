package cmd

import (
	"testing"
)

type mockPr struct {
	title string
}

func (pr mockPr) Title() string {
	return pr.title
}

// I'm not smart enough to write good test cases, that's why I just dump a bunch of exmaples with expected outcome here.
//
// I use this for exploratory testing and development by looking at existing release changelogs of Terasology or the
// TerasologyLauncher and just dump the respective title with the associated category here.
//
// To improve this, one could derive assumptions and rules that should hold for the category derivation.
// This could be, for instance:
//  - known prefixes are matched correctly if the title is in format `^\w+[\([a-zA-Z0-9\-]*\)]?:.*`
//  - if not matching this format, the PR is matched to the GENERAL category
//  - unknown prefixes are matched to the GENERAL category
//  - known prefixes appearing somewhere else within the title don't affect the matching
//	- matching for known prefixes should be case insensitive
//  - (potential) non-exact matches, e.g., "fixes: " instead of "fix", or common typos
//
// If anybody has an idea how to structure this better please let me know or feel free to raise a PR.
func TestGetPrCategory(t *testing.T) {
	cases := []struct {
		title    string
		expected PrCategory
		name     string
	}{
		{"feat: Update engine settings i18n", FEATURES, "short prefix"},
		{"feat(i18n): Update engine settings i18n", FEATURES, "short prefix with scope"},
		{"feature:  Update engine settings i18n", FEATURES, "long prefix"},
		{"feature(i18n):  Update engine settings i18n", FEATURES, "long prefix with scope"},

		{"fix: use Maps from guava, not Google API Client", BUG_FIXES, "prefix"},
		{"fix(build): use Maps from guava, not Google API Client", BUG_FIXES, "prefix with scope"},
		{"bugfix: use Maps from guava, not Google API Client", BUG_FIXES, "prefix variant"},
		{"fixes: use Maps from guava, not Google API Client", BUG_FIXES, "prefix variant"},
		{"fixed: use Maps from guava, not Google API Client", BUG_FIXES, "prefix variant"},

		{"feat(foo) my cool feature", GENERAL, "prefix with scope, but no separator"},
		{"foo(bar): feat: X is broken - this fixes it", GENERAL, "unknown prefix / contains valid prefix"},
		{"feature X is broken - this fixes it", GENERAL, "misleading prefix (no separator)"},
		{"Fixed checkstyle issues", GENERAL, "misleading prefix (no separator)"},

		{"chore: use picocli for processing command line options", MAINTENANCE, "prefix"},
		{"chore(facade): use picocli for processing command line options", MAINTENANCE, "prefix with scope"},
		{"chore[facade]: use picocli for processing command line options", MAINTENANCE, "prefix with scope variant"},
		{"refactor: transaction manager with reactor", MAINTENANCE, "prefix variant"},
		{"refactor(reactor): transaction manager with reactor", MAINTENANCE, "prefix variant with scope"},

		{"doc: update minimal system requirements for OpenGL 3.3", DOCUMENTATION, "prefix"},
		{"docs: update minimal system requirements for OpenGL 3.3", DOCUMENTATION, "prefix variant"},
		{"documentation: update minimal system requirements for OpenGL 3.3", DOCUMENTATION, "prefix variant"},

		{"build: build using a java 11 toolchain", LOGISTICS, "prefix"},
		{"build(ci): build using a java 11 toolchain", LOGISTICS, "prefix with scope"},
		{"ci: build using a java 11 toolchain", LOGISTICS, "prefix variant"},

		{"perf: upgrade to use proto3", PERFORMANCE, "prefix"},
		{"perf(serialization): upgrade to use proto3", PERFORMANCE, "prefix with scope"},
		{"performance: upgrade to use proto3", PERFORMANCE, "prefix variant"},

		{"test: Convert to MTEExtension", TESTS, "prefix"},
		{"test(mte): Convert to MTEExtension", TESTS, "prefix with scope"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			category := getPrCategory(mockPr{testCase.title})
			if testCase.expected != category {
				t.Errorf("Derived category of '%s' was incorrect. got: %s, want: %s.", testCase.title, category.String(), testCase.expected.String())
			}
		})
	}
}
