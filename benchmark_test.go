package loads

import (
	_ "embed"
	"encoding/json"
	"strconv"
	"testing"
)

//go:embed fixtures/json/bench/header.partial
var benchHeader []byte

//go:embed fixtures/json/bench/path-item.partial
var benchPathItem []byte

//go:embed fixtures/json/bench/footer.partial
var benchFooter []byte

func BenchmarkAnalyzed(b *testing.B) {
	d := make([]byte, 0, len(benchHeader)+1000*(len(benchPathItem)+20)+len(benchFooter))
	d = append(d, benchHeader...)

	for i := range 1000 {
		d = append(d, `,
    "/pets/`...)
		d = strconv.AppendInt(d, int64(i), 10)
		d = append(d, `": `...)
		d = append(d, benchPathItem...)
	}

	d = append(d, benchFooter...)
	rm := json.RawMessage(d)
	b.ResetTimer()

	for b.Loop() {
		_, err := Analyzed(rm, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}
