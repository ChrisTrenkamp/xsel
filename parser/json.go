package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/ChrisTrenkamp/xsel/node"
)

type JsonElement struct {
	local string
}

func (j JsonElement) Space() string {
	return ""
}

func (j JsonElement) Local() string {
	return j.local
}

type JsonCharData struct {
	value string
}

func (j JsonCharData) CharDataValue() string {
	return j.value
}

type jsonState int

const (
	DEFAULT jsonState = iota
	READ_OBJECT_FIELD
	READ_OBJECT_VALUE
	EMIT_OBJECT_VALUE
	OBJECT_VALUE_EMITTED
	READ_ARRAY
	READ_ARRAY_VALUE
	ARRAY_VALUE_EMITTED
)

type jsonParser struct {
	jsonReader  *json.Decoder
	nameStack   []string
	stateStack  []jsonState
	stagedToken json.Token
}

func (j *jsonParser) appendState(s jsonState) {
	j.stateStack = append(j.stateStack, s)
}

func (j *jsonParser) replaceState(s jsonState) {
	j.popState()
	j.appendState(s)
}

func (j *jsonParser) currentState() jsonState {
	if len(j.stateStack) == 0 {
		return DEFAULT
	}

	return j.stateStack[len(j.stateStack)-1]
}

func (j *jsonParser) popState() {
	if len(j.stateStack) > 0 {
		j.stateStack = j.stateStack[:len(j.stateStack)-1]
	}
}

func (j *jsonParser) appendName(s string) {
	j.nameStack = append(j.nameStack, s)
}

func (j *jsonParser) currentName() string {
	if len(j.nameStack) == 0 {
		return ""
	}

	return j.nameStack[len(j.nameStack)-1]
}

func (j *jsonParser) popName() {
	if len(j.nameStack) > 0 {
		j.nameStack = j.nameStack[:len(j.nameStack)-1]
	}
}

func (j *jsonParser) Pull() (node.Node, bool, error) {
	if j.currentState() == EMIT_OBJECT_VALUE {
		j.replaceState(OBJECT_VALUE_EMITTED)
		return jsonTokenValue(j.stagedToken), false, nil
	}

	if j.currentState() == OBJECT_VALUE_EMITTED {
		j.replaceState(READ_OBJECT_FIELD)
		j.popName()
		return nil, true, nil
	}

	if j.currentState() == READ_ARRAY_VALUE {
		j.replaceState(ARRAY_VALUE_EMITTED)
		return jsonTokenValue(j.stagedToken), false, nil
	}

	if j.currentState() == ARRAY_VALUE_EMITTED {
		j.replaceState(READ_ARRAY)
		return nil, true, nil
	}

	tok, err := j.jsonReader.Token()

	if err != nil {
		return nil, false, err
	}

	if j.currentState() == READ_OBJECT_VALUE {
		switch tok.(type) {
		case json.Delim:
		default:
			j.replaceState(EMIT_OBJECT_VALUE)
			j.stagedToken = tok
			return JsonElement{local: j.currentName()}, false, nil
		}
	}

	if j.currentState() == READ_ARRAY {
		switch tok.(type) {
		case json.Delim:
		default:
			j.replaceState(READ_ARRAY_VALUE)
			j.stagedToken = tok
			return JsonElement{local: j.currentName()}, false, nil
		}
	}

	switch t := tok.(type) {
	case json.Delim:
		return parseJsonDelim(j, t)
	case string:
		if j.currentState() == READ_OBJECT_FIELD {
			j.replaceState(READ_OBJECT_VALUE)
			j.appendName(t)
		}
	}

	return j.Pull()
}

func jsonTokenValue(tok json.Token) node.Node {
	switch t := tok.(type) {
	case bool:
		return JsonCharData{value: fmt.Sprintf("%t", t)}
	case float64:
		str := strconv.FormatFloat(t, 'g', -1, 64)
		return JsonCharData{value: str}
	case json.Number:
		return JsonCharData{value: string(t)}
	case string:
		return JsonCharData{value: t}
	}

	return JsonCharData{value: "null"}
}

func parseJsonDelim(j *jsonParser, t json.Delim) (node.Node, bool, error) {
	switch t.String() {
	case "{":
		state := j.currentState()
		j.appendState(READ_OBJECT_FIELD)

		if state == READ_OBJECT_VALUE || state == READ_ARRAY {
			return JsonElement{local: j.currentName()}, false, nil
		}
	case "}":
		j.popState()

		if j.currentState() == READ_OBJECT_VALUE {
			j.replaceState(READ_OBJECT_FIELD)
			return nil, true, nil
		}

		if j.currentState() == READ_ARRAY {
			return nil, true, nil
		}
	case "[":
		j.appendState(READ_ARRAY)
	case "]":
		j.popState()

		if j.currentState() == READ_OBJECT_VALUE {
			j.popName()
			j.replaceState(READ_OBJECT_FIELD)
		}
	}

	return j.Pull()
}

// Create a Parser that reads the given JSON document.
func ReadJson(in io.Reader) Parser {
	jsonReader := json.NewDecoder(in)

	return &jsonParser{
		jsonReader: jsonReader,
		nameStack:  make([]string, 0),
		stateStack: make([]jsonState, 0),
	}
}
