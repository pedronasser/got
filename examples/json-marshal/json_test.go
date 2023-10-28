//go:build !generated

package json_marshal

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func BenchmarkGoJSONMarshal(b *testing.B) {
	f, err := os.OpenFile("user.json", os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	var user User

	err = json.NewDecoder(f).Decode(&user)
	if err != nil {
		panic(err)
	}

	sample, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(sample))

	var users = make([]User, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = user
		users[i].Id = i
	}

	for i := 0; i < b.N; i++ {
		for _, u := range users {
			json.Marshal(u)
		}
	}
}
