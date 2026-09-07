package secrets

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync/atomic"
)

// agentsviewTestFixtureHashes holds SHA-256 hashes for secret-shaped strings
// that appear in agentsview's own unit tests, including fixtures built from
// split string fragments. The raw values are intentionally not stored here:
// committing a deny-list of literal tokens creates another source file that a
// recorded agent session can quote and then report as a leak.
//
// Adding a new positive-path secret fixture to a test? Add only its SHA-256
// hash here. The fixture hash list is folded into RulesVersion, so backfill
// treats deny-list changes as scanner changes and re-scans stale sessions.
var agentsviewTestFixtureHashes = map[string]struct{}{
	"00af946788bf19727386a232e4058d5038f40847770182d95643fca350d996e6": {},
	"039a47e1ca7c07d6514be7a9a51c185e6b0e276a2762ed6d8df0a134d36c3391": {},
	"05a943eac035fd10c9e844015aaab9f4c329f175d855154c9188b2a125279dd1": {},
	"08d4c233406d9209ddba77a3fbaa6be61a74dc1e4fa6e10a2b449c146256f551": {},
	"0184e1dcd360ac23565ae9cc057e0b91e3574449fd0f8862526a43936d13977b": {},
	"0f604417a7862a3da3c7741529beefe8fdfb98042da86a6b8d4da27ff3098c1a": {},
	"17c22292f944fcaba03e4bd70bc0871464590bd9eb47930fbc2a91a886c69346": {},
	"1b99c0d6422aa645d3562e482203704ca1d9114e5bffaf840dfeba9aadeff156": {},
	"20f597970aa2fe4a781204b5105389b0868765520c671c8e8f440ccc2907cf94": {},
	"2c2c3f90a9afeb084629366c3a5a67b3e9fce9373744c1b971ae77f1b617b2e6": {},
	"35bb9f766ebd02dbbb021f53e3adf97da69d342b1b6c9fb8618c525829b1037a": {},
	"378a8d661936c5c4f5bc20fbcb6ad7395e00c95fd059d4cceb2d56d54eaccff5": {},
	"531660540c834c1b12e79db2fc9936411648e234487c9000a89941d8c24f69ee": {},
	"569179513f9c76b5d1354200ffc38b24fc10bc7ee356a0ab8c9553433ca79745": {},
	"5731daff90361fdfd01208e00a0e34b7b5905d7088d62c8349882d5c4c530e69": {},
	"59304de2f7780c2ca6dcdeafa13f16ad6d11b3f249f118c76eede16a4dabddc2": {},
	"5ab0ac8280808a9394f7584cf30af03799f29437d53626ac74be69029f5c1dd5": {},
	"5d6604e3ba794f08dd0900672aa06242029fa3dbc90725b9c1f6fb261045f443": {},
	"68f13b41b6aa48ecb511ac1ef2ccb29d172b71a22b0a5bbbf2a63d387b5ac423": {},
	"743554670c6065b3f7f13ac4f07e392f977b3556ceb7457411633c454bcbece8": {},
	"8cc9a7f047d27eb9c2a27b879ee6dea7a1b06f58b47802aa2887af29013715ee": {},
	"9138fd25248a002f43de54f55feb57182a56a43e307ff31b878d4822f0b2120d": {},
	"9383d3a321ec3ee35cc104a887d56aa8cd3543fd29eae27b40cebc14354bfc2e": {},
	"964f175c0bbd3141f95c49fc9b0d0fb7953921bc28c45480ed88d0e48f4c3f05": {},
	"9a0b5b544911419929246614f134083d67d6e58fb2eed48cbf14d349ec3fc1e6": {},
	"9a1015d45e55e0daaf7020d79adffa8987cc4c40d838797b13bd86f28faf8f74": {},
	"9a7ec609dc2b71a06470d9af744fab4bdc8040288adc562845d54a01f3ca36c8": {},
	"a82ed2d14be075cf2af62de3911f8fade1b12182dc81af18b9fa97ba6b0457ff": {},
	"af8207cf94138ad22f72a35b81ee53668fe8d3d14e21cf4aa29f7e5c5e12cddd": {},
	"b2d1ca1a67caf3326751b12ef347a5c525d680319a0513ff836398c3182242a7": {},
	"b9e542bb492ae97fb2006d797da5b389a1237674d713c49eabc1ff6d99bec67c": {},
	"bd0fdfb6f21d7bfcec3c7086c9c86fb286d4f020d97d9f8df4fa15b4ef6b4d3d": {},
	"c14f984e31e231914767b815b1e9c71acd3825cb1c20bda419b39a0d60db9eb5": {},
	"c7d894bcb8b6ae7416f2aa6451723d1ac7b17c999edbe03849f63159e8e2f53a": {},
	"ca3533eaa36311ec950e1443330a7b0b4a452f247f4f478f419a469a95d2cae8": {},
	"cfae9a6eb09ea990634453c17ad47f9b3da587dd78e0e5db60238802f2efbedb": {},
	"d3edb1d523718e22784ca5f4d7b033ebe18208b823f3c8e6c8ab2028148199a6": {},
	"ddde2240b099bf523beab304d905a2b62490a2c9169ba7eb374210805d56653c": {},
	"e38b3c3f77755fd8f5d4dd1c8e9f305313ea15ab8144432e08ac4b1a08f631d6": {},
	"e40d52868dd5a755e22a235de77b057bb08e7131a7e7a895ad6468d420bc3d9a": {},
	"fbe68dd1d502fded62be57881c10ea10fdf5205f20bbaeff8096e1e3c2e2309a": {},
}

// isAgentsviewTestFixture reports whether s is one of agentsview's known
// secret-shaped test fixtures.
func isAgentsviewTestFixture(s string) bool {
	_, ok := agentsviewTestFixtureHashes[fixtureHash(s)]
	return ok
}

func fixtureHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func sortedAgentsviewFixtureHashes() []string {
	out := make([]string, 0, len(agentsviewTestFixtureHashes))
	for h := range agentsviewTestFixtureHashes {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// fixtureDenyEnabled controls whether Scan filters matches against
// agentsviewTestFixtureHashes. False by default so unit tests (which build
// their input around random-looking fixtures the rules need to verify positive
// paths against) pass without per-test boilerplate. The agentsview binary calls
// EnableFixtureDeny at startup so production scans automatically suppress
// agentsview's own fixture noise.
var fixtureDenyEnabled atomic.Bool

// EnableFixtureDeny turns on the agentsview-test-fixture deny-list for
// subsequent Scan and ScanDefinite calls. Wired into the CLI entrypoint so the
// long-running server, ad-hoc CLI commands, and sync engine all filter fixture
// noise. Off by default so unit tests can assert positive rule paths against
// the same values.
func EnableFixtureDeny() {
	fixtureDenyEnabled.Store(true)
}

// disableFixtureDenyForTest restores fixtureDenyEnabled to its previous value
// after the cleanup runs. Used by the secrets package own tests that want to
// exercise the deny-list path explicitly.
func disableFixtureDenyForTest(cleanup func(func())) {
	prev := fixtureDenyEnabled.Swap(false)
	cleanup(func() { fixtureDenyEnabled.Store(prev) })
}
