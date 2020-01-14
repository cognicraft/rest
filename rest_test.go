package rest

import "testing"

func TestParser(t *testing.T) {
	r, err := ParseOneString(batch)
	if err != nil {
		t.Error(err)
	}
	DumpRequest(r)

	for _, r := range ParseAllString(batch) {
		DumpRequest(r)
	}

	// t.Fail()
}

const single = `
POST http://example.com/baz/ HTTP/1.1
Content-Type: application/json

{
    "foo": "bar"
}
`

const batch = `
POST http://example.com/baz/ HTTP/1.1
Content-Type: application/json

{
    "foo": "bar"
}

@host = localhost:8080

POST http://{{host}}/baz/ HTTP/1.1
Content-Type: application/json

{
	"foo": "bar",
	"ts": "{{$time}}",
	"berlin-ts": "{{$time Europe/Berlin}}",
	"berlin-ts-other": "{{$time Europe/Berlin dd/MM/YYYY}}",
	"id": "{{$uuid}}",
	"other-id": "{{$uuid short}}"
}
`
