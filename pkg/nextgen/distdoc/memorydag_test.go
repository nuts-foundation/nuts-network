package distdoc

import "testing"

var memoryDAGCreator = func(_ *testing.T) DAG {
	return NewMemoryDAG()
}

func TestMemoryDAG_Add(t *testing.T) {
	DAGTest_Add(memoryDAGCreator, t)
}

func TestMemoryDAG_MissingDocuments(t *testing.T) {
	DAGTest_MissingDocuments(memoryDAGCreator, t)
}

func TestMemoryDAG_Walk(t *testing.T) {
	DAGTest_Walk(memoryDAGCreator, t)
}
