package utils_test

import (
	"sync"
	"testing"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestAppendSorted_Ints(t *testing.T) {
	valuesToSort := []int{1, 3, 2, 5, 4, 6}

	values := []int{}

	for _, v := range valuesToSort {
		values = utils.AppendSorted(values, v, func(i, j int) bool {
			return i > j
		})
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, values)
}

type testStruct struct {
	Name string
}

func TestAppendSorted_StructPointers(t *testing.T) {
	valuesToSort := []*testStruct{
		{Name: "a"},
		{Name: "c"},
		{Name: "b"},
		{Name: "e"},
		{Name: "d"},
	}

	values := []*testStruct{}

	for _, v := range valuesToSort {
		values = utils.AppendSorted(values, v, func(i, j *testStruct) bool {
			return i.Name > j.Name
		})
	}

	assert.Equal(t, []*testStruct{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
		{Name: "d"},
		{Name: "e"},
	}, values)
}

func TestToBool_ConfigMap(t *testing.T) {}

func TestDedent(t *testing.T) {
	assert.Equal(t, "test", utils.Dedent(`test`))
	assert.Equal(t, "test", utils.Dedent(`
		test
	`))
	assert.Equal(t, `test
		test`, utils.Dedent(`test
		test
	`))
	assert.Equal(t, `test
test`, utils.Dedent(`
		test
		test
	`))
	assert.Equal(t, `test
test
	test
		test
	test
test`, utils.Dedent(`
		test
		test
			test
				test
			test
		test
	`))
	assert.Equal(t, `test
	test
		test`, utils.Dedent(`
		test
			test
				test
	`))
	assert.NotEqual(t, `test
	test
		test`, utils.Dedent(`test`))
}

func TestOneManQueue_SerialExecution(t *testing.T) {
	var mu sync.Mutex
	executionOrder := []int{}

	id := 0

	fn := func() {
		mu.Lock()
		executionOrder = append(executionOrder, id)
		mu.Unlock()
		// Simulate work
		time.Sleep(50 * time.Millisecond)
	}

	enqueue := utils.OneManQueue(fn)

	var wg sync.WaitGroup
	numCalls := 5
	wg.Add(numCalls)

	for i := 0; i < numCalls; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			id++
			mu.Unlock()
			enqueue()
		}()
	}

	wg.Wait()

	mu.Lock()
	if len(executionOrder) != 2 {
		t.Errorf("Expected executionOrder length 2, got %d", len(executionOrder))
	}

	if executionOrder[0] != 1 || executionOrder[1] != 5 {
		t.Errorf("Expected executionOrder to be [1, 5], got %v", executionOrder)
	}
	mu.Unlock()

	id++
	enqueue()

	if len(executionOrder) != 3 {
		t.Errorf("Expected executionOrder length 3, got %d", len(executionOrder))
	}

	if executionOrder[2] != 6 {
		t.Errorf("Expected executionOrder to be [1, 5, 6], got %v", executionOrder)
	}
}
