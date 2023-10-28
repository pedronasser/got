package transform

import (
	"bytes"
	"testing"
)

func testExtractAttributeUsage(t *testing.T, testcase string, expected []string) {
	usages, err := extractAttributeUsages(bytes.NewReader([]byte(testcase)))
	if err != nil {
		t.Fatal(err)
	}

	index := 0
	for _, usage := range usages {
		for _, attr := range usage.attributes {
			if attr.Name != expected[index] {
				t.Errorf("Expected %s, got %s", expected[index], attr.Name)
			}
			index++
		}
	}
}

func TestExtractAttributeUsages(t *testing.T) {
	testExtractAttributeUsage(t,
		`//go:build generated
package test

//Normal comment
//Another normal comment

//#[Foo]
type Bar struct{}

func main() {
	// More comments
	// More comments
	// More comments

	//#[Foo]
	fmt.Println("hello")
}

// #[decorator]
func Foo() {}`,
		[]string{"Foo", "Foo", "decorator"},
	)
}
