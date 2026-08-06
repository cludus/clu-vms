package utils

import (
	"testing"
)

type Person struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func TestParseYaml(t *testing.T) {
	content := `
name: Alice
age: 30
`
	var person Person
	err := ParseYaml(content, &person)
	if err != nil {
		t.Fatalf("ParseYaml returned error: %v", err)
	}
	if person.Name != "Alice" {
		t.Errorf("expected Name to be Alice, got %q", person.Name)
	}
	if person.Age != 30 {
		t.Errorf("expected Age to be 30, got %d", person.Age)
	}
}

func TestSerializeYaml(t *testing.T) {
	person := Person{Name: "Bob", Age: 25}
	out, err := SerializeYaml(person)
	if err != nil {
		t.Fatalf("SerializeYaml returned error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty YAML output")
	}
}

func TestParseAndSerializeYaml(t *testing.T) {
	original := Person{Name: "Charlie", Age: 40}
	serialized, err := SerializeYaml(original)
	if err != nil {
		t.Fatalf("SerializeYaml returned error: %v", err)
	}
	var parsed Person
	if err := ParseYaml(serialized, &parsed); err != nil {
		t.Fatalf("ParseYaml returned error: %v", err)
	}
	if parsed != original {
		t.Errorf("expected parsed person %+v to equal original %+v", parsed, original)
	}
}

type Team struct {
	Name    string   `yaml:"name"`
	Lead    Person   `yaml:"lead"`
	Members []Person `yaml:"members"`
}

func TestParseYamlWithNestedPerson(t *testing.T) {
	content := `
name: DevOps
lead:
  name: Dana
  age: 35
members:
  - name: Alice
    age: 30
  - name: Bob
    age: 25
`
	var team Team
	err := ParseYaml(content, &team)
	if err != nil {
		t.Fatalf("ParseYaml returned error: %v", err)
	}
	if team.Name != "DevOps" {
		t.Errorf("expected Team Name to be DevOps, got %q", team.Name)
	}
	if team.Lead != (Person{Name: "Dana", Age: 35}) {
		t.Errorf("expected Lead %+v, got %+v", Person{Name: "Dana", Age: 35}, team.Lead)
	}
	if len(team.Members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(team.Members))
	}
	if team.Members[0] != (Person{Name: "Alice", Age: 30}) {
		t.Errorf("expected first member %+v, got %+v", Person{Name: "Alice", Age: 30}, team.Members[0])
	}
	if team.Members[1] != (Person{Name: "Bob", Age: 25}) {
		t.Errorf("expected second member %+v, got %+v", Person{Name: "Bob", Age: 25}, team.Members[1])
	}
}
