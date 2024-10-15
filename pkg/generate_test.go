package pkg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockOperation struct {
	Foo string
}

type MockOperationState struct {
	Bar string `json:"bar"`
}

func (m *MockOperation) Generate(state string) (interface{}, error) {
	var mockState MockOperationState
	err := json.Unmarshal([]byte(state), &mockState)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"foo": m.Foo,
		"bar": mockState.Bar,
	}, nil
}

func (m *MockOperation) Name() OperationName {
	return "test"
}

func (m *MockOperation) IsAsync() bool {
	return false
}

func (m *MockOperation) Update(string) (string, error) {
	return "", nil
}

type MockOperation2 struct {
	Baz string
}

func (m *MockOperation2) Generate(state string) (interface{}, error) {
	return map[string]string{
		"baz": m.Baz,
	}, nil
}

func (m *MockOperation2) Name() OperationName {
	return "test2"
}

func (m *MockOperation2) IsAsync() bool {
	return false
}

func (m *MockOperation2) Update(string) (string, error) {
	return "", nil
}

func TestGenerateContent(t *testing.T) {
	config := AllConfigs{
		Configs: map[LocationKey]Location{
			"pane": {
				Operations: []OperationWrapper{
					{
						Operation: &MockOperation{
							Foo: "foo",
						},
					},
					{
						Operation: &MockOperation2{
							Baz: "baz",
						},
					},
				},
				Template: "test > {{ .test.foo }} > {{ .test.bar }} > {{ .test2.baz }}",
			},
		},
	}

	memoryStateStore := NewMemoryStateStore()

	memoryStateStore.Set(LocationKey("pane"), InstanceKey("test"), OperationName("test"), `{"bar": "bar"}`)
	content, err := GenerateContent(memoryStateStore, config.Configs["pane"], LocationKey("pane"), InstanceKey("test"))
	require.NoError(t, err)
	require.Equal(t, "test > foo > bar > baz", content)
}
