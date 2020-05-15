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

func TestSetFailures(t *testing.T) {
	setTests := []struct {
		source  string
		patch   string
		wantErr string
	}{
		{
			source:  ": testing\n",
			patch:   "name",
			wantErr: "yaml: did not find expected key",
		},
		{
			source:  "name: testing\n",
			patch:   "",
			wantErr: "path cannot be empty",
		},
	}

	for i, tt := range setTests {
		_, err := SetBytes([]byte(tt.source), tt.patch, "testing")
		if err.Error() != tt.wantErr {
			t.Fatalf("%d failed, got %s, want %s", i, err, tt.wantErr)
		}
	}
}
