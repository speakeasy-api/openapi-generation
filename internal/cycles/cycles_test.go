package cycles

import (
	"sort"
	"strconv"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/assert"
)

// Some test cases taken from https://github.com/williamfiset/Algorithms/blob/master/src/test/java/com/williamfiset/algorithms/graphtheory/TarjanSccSolverAdjacencyListTest.java

func createGraph(n int) ast.TypeDefs {
	types := make(ast.TypeDefs, n)
	for i := 0; i < n; i++ {
		types[i] = &ast.TypeDef{
			Name: strconv.Itoa(i),
		}
	}
	return types
}

func addEdge(g ast.TypeDefs, i, j int) {
	g[i].Fields = append(g[i].Fields, &ast.FieldDef{
		Type: g[j],
	})
}

func TestSingletonCase(t *testing.T) {
	g := createGraph(1)
	actual := collectCycles(t, g)
	assert.Equal(t, [][]int{}, actual)
}

func TestAreCircular_TwoDisjointCycles(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpNj7sOwjAMRX8l8lyklrLQgQk2WIAJebEa9yERp0oTiYf4d0IqKJ6Ofe1r3SfUVjNU0DoaOrU_ouRqsdioAqVIkKMsE5QoZYIVyirBEjIw7Az1Ojo8UZRC8B0bRqgiam4oXD1CNkmGbjvd8vhRi_yvZv3MN3_qH8lgPS-gvOIvCt6e7lJD5V3gDJwNbQdVQ9cxdmHQ5HnbU4xiftOB5GKt-Z6w7r11hylziv56A0a9Tec
	g := createGraph(5)
	addEdge(g, 0, 1)
	addEdge(g, 1, 0)
	addEdge(g, 2, 3)
	addEdge(g, 3, 4)
	addEdge(g, 4, 2)

	expected := [][]int{{0, 1}, {2, 3, 4}}
	actual := collectCycles(t, g)

	assert.Equal(t, expected, actual)
}

func TestButterflyCase(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpVkLsOwjAMRX8l8lyklrLQgQk2WCgT8mI1pq1EnCpNJB7qvxOCeHk69rWvZd-hsZqhgtbR0KntHiVXs9lKFShFgjnKPEGJUv5LC5RFghwyMOwM9Tpa3VGUQvAdG0aoImo-UTh7hOwlGbpsdMvjUy3yn_jqB774ur8lg-W3AWWKuyh4W1-lgcq7wBk4G9oOqhOdx5iFQZPndU_xJvOpDiRHa817hHXvrdu9jk8_mB5sSE_r
	g := createGraph(5)
	addEdge(g, 0, 1)
	addEdge(g, 1, 2)
	addEdge(g, 2, 3)
	addEdge(g, 3, 1)
	addEdge(g, 1, 4)
	addEdge(g, 4, 0)

	expected := [][]int{{0, 1, 2, 3, 4}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestDisjointTree(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpNj70OwjAMhF8l8txKlD-JDEywwUKZkBeLuG0kklQhkQoV705oBcXTZ5991vVwdYpBQu2pbcThhHYm8nwrCrTFAHO0ywEWaFdfWI8AGRj2hrRKDj1aIRBCw4YRZELFFcVbQMhGyVC3VzXfP2ox-6tJP3MXSv0cDDbTAtpX-kUxuPJhryCDj5yBd7FuQFZ0u6cutooC7zSlKOY3bclenDPfE1Y6OH8cMw_RX29NdE3v
	g := createGraph(7)
	addEdge(g, 0, 1)
	addEdge(g, 1, 2)
	addEdge(g, 4, 3)
	addEdge(g, 5, 3)
	addEdge(g, 6, 3)

	expected := [][]int{}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestDisjointTreeFromHackerrank(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpdkDFvwjAQhf-KdXOQcKC0ycDUbnQpnapbTvhIIuFzZGyJFvHfa2KSVPX0-b3zO-td4eAMQw2Np75Vuw-UlVostkqj6DLTHZcDVigvAzwlKZvlZJaTtkZ5Hucfpl4l1GPy5l9IonlXpg1K9ZCgAMveUmfSP68oSiGEli0j1AkNHymeAkKRLUuXN9Pw-e7q5Z8z-598CfvuZwio5gGUW9pFMbj9txygDj5yAd7FpoX6SKdzusXeUODXjlJhdlJ7ki_n7PiETRecf8_NDgXffgGQ3mSV
	g := createGraph(16)
	addEdge(g, 3, 1)
	addEdge(g, 12, 11)
	addEdge(g, 10, 9)
	addEdge(g, 8, 5)
	addEdge(g, 1, 12)
	addEdge(g, 10, 2)
	addEdge(g, 1, 14)
	addEdge(g, 7, 9)
	addEdge(g, 10, 13)
	addEdge(g, 11, 1)
	addEdge(g, 6, 5)
	addEdge(g, 1, 15)
	addEdge(g, 2, 11)
	addEdge(g, 2, 6)
	addEdge(g, 9, 11)

	expected := [][]int{{1, 11, 12}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestFirstGraphInSlides(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpNkMFuwjAMhl8l8rlIbVmB9bDTuLHL4DT5EhHTViJOFRKJgXj3ZYk8ltOX73fsyHc4OkPQw-D1PKrdJ3KtFos31SA3GWoxG-SNmAJr5HWGFfJKTCPQSp9WajqBYjoxS-Rlhhb5RUyBDiqw5K2eTPrkHVkphDCSJYQ-oaGTjueAUJXI6uvWDHT5TZv633nmB7qG_XTLDV6fBciPNEvH4PbffIQ--EgVeBeHEfqTPl_SLc5GB3qfdNqW_bOz5i_nrDwhMwXnP8pa83YfPx2ZZGY
	g := createGraph(9)
	addEdge(g, 0, 1)
	addEdge(g, 1, 0)
	addEdge(g, 0, 8)
	addEdge(g, 8, 0)
	addEdge(g, 8, 7)
	addEdge(g, 7, 6)
	addEdge(g, 6, 7)
	addEdge(g, 1, 7)
	addEdge(g, 2, 1)
	addEdge(g, 2, 6)
	addEdge(g, 5, 6)
	addEdge(g, 2, 5)
	addEdge(g, 5, 3)
	addEdge(g, 3, 2)
	addEdge(g, 4, 3)
	addEdge(g, 4, 5)

	expected := [][]int{{0, 1, 8}, {2, 3, 5}, {6, 7}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestLastGraphInSlides(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpN0D0PwiAQBuC_0txck1bbGjs46aaLOhmWi5xtk1IaCokf8b-LXFRu4uHlgNwTLloS1NAYHNtkdxBD4itLZrN1kjPygDljHpAxFgFFjCWjCCgZZdzDqBhVnFTxO1V89TJgEaOEFBQZhZ30v39-IgG2JUUCar-UdEXXWwEpRwpvW9nQ9EnzLKp_fqKbPXaPcMHqf0AML_8WOquP9-ECtTWOUjDaNS3UV-wnLzdKtLTp0I9R_XZHHM5aq28Lyc5qs-d5h7G_3p8HZME
	g := createGraph(8)
	addEdge(g, 0, 1)
	addEdge(g, 1, 2)
	addEdge(g, 2, 0)
	addEdge(g, 3, 4)
	addEdge(g, 3, 7)
	addEdge(g, 4, 5)
	addEdge(g, 5, 0)
	addEdge(g, 5, 6)
	addEdge(g, 6, 0)
	addEdge(g, 6, 2)
	addEdge(g, 6, 4)
	addEdge(g, 7, 3)
	addEdge(g, 7, 5)

	expected := [][]int{{0, 1, 2}, {3, 7}, {4, 5, 6}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestCycleWithBidirectionalBridge(t *testing.T) {
	// https://mermaid.live/edit#pako:eNpNkLsOwjAMRX8l8txKLYWBDkywwUKZkBeLmLYScaqQSDzEvxNSFfB07Gtfy37CyWqGGlpHQ6e2e5RC5flKlShlghnKLEExQYVSTdIIc5R5ggXKYuyBDAw7Q72O5k8UpRB8x4YR6oiazxQuHiEbJUO3jW75-lHL4i9--oFvvukfyWD5a0B5xV0UvG3ucoLau8AZOBvaDuozXa4xC4Mmz-ue4pXmWx1IjtaaaYR1763bje9IX3m9AfjIVAE
	g := createGraph(6)
	addEdge(g, 0, 1)
	addEdge(g, 1, 2)
	addEdge(g, 2, 0)
	addEdge(g, 2, 3)
	addEdge(g, 3, 2)
	addEdge(g, 3, 4)
	addEdge(g, 4, 5)
	addEdge(g, 5, 3)

	expected := [][]int{{0, 1, 2, 3, 4, 5}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestSelfReferentialSingleNode(t *testing.T) {
	g := createGraph(1)
	addEdge(g, 0, 0)

	expected := [][]int{{0}}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestNotACycleStraightLine(t *testing.T) {
	g := createGraph(8)
	addEdge(g, 0, 1)
	addEdge(g, 1, 2)
	addEdge(g, 2, 3)
	addEdge(g, 4, 5)
	addEdge(g, 5, 6)
	addEdge(g, 6, 7)

	expected := [][]int{}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestNotACycleFork(t *testing.T) {
	g := createGraph(3)
	addEdge(g, 0, 1)
	addEdge(g, 0, 2)
	addEdge(g, 2, 1)

	expected := [][]int{}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func TestNotACycleSingleNode(t *testing.T) {
	g := createGraph(1)

	expected := [][]int{}
	actual := collectCycles(t, g)
	assert.Equal(t, expected, actual)
}

func collectCycles(t *testing.T, g ast.TypeDefs) [][]int {
	t.Helper()
	detector := New()
	detector.DetectCycles(g)
	cycles := make([][]int, len(detector.Cycles))
	for i, cycle := range detector.Cycles {
		cycles[i] = []int{}
		for node := range cycle {
			cycles[i] = append(cycles[i], utils.Must(strconv.Atoi(node.Name)))
		}
		sort.Ints(cycles[i])
	}
	sort.Slice(cycles, func(i, j int) bool {
		return cycles[i][0] < cycles[j][0]
	})
	return cycles
}
