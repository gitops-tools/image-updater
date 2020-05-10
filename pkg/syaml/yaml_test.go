package syaml

import (
	"testing"
)

func TestSet(t *testing.T) {
	setTests := []struct {
		source   string
		patch    string
		newValue interface{}
		want     string
	}{
		{
			source:   "name: testing\n",
			patch:    "name",
			newValue: "new name",
			want:     "name: new name\n",
		},
		{
			source:   "person:\n  age: 30\n  name: John\n",
			patch:    "person.name",
			newValue: "Anderson",
			want:     "person:\n  age: 30\n  name: Anderson\n",
		},
		{
			source:   "items:\n- age: 30\n- age: 29\n",
			patch:    "items.1.age",
			newValue: 20,
			want:     "items:\n- age: 30\n- age: 20\n",
		},
	}

	for i, tt := range setTests {
		updated, err := SetBytes([]byte(tt.source), tt.patch, tt.newValue)
		if err != nil {
			t.Error(err)
			continue
		}

		if string(updated) != tt.want {
			t.Errorf("%d failed, got %#v, want %#v", i, string(updated), tt.want)
		}
	}
}
