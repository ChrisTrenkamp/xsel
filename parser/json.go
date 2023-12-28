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
	defaultJsonState jsonState = iota
	readObjectFieldState
	readObjectValueState
	emiteObjectValueState
	objectValueEmittedState
	readArrayState
	readArrayValueState
	arrayValueEmittedState
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
		return defaultJsonState
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
	if j.currentState() == emiteObjectValueState {
		j.replaceState(objectValueEmittedState)
		return jsonTokenValue(j.stagedToken), false, nil
	}

	if j.currentState() == objectValueEmittedState {
		j.replaceState(readObjectFieldState)
		j.popName()
		return nil, true, nil
	}

	if j.currentState() == readArrayValueState {
		j.replaceState(arrayValueEmittedState)
		return jsonTokenValue(j.stagedToken), false, nil
	}

	if j.currentState() == arrayValueEmittedState {
		j.replaceState(readArrayState)
		return nil, true, nil
	}

	tok, err := j.jsonReader.Token()

	if err != nil {
		return nil, false, err
	}

	if j.currentState() == readObjectValueState {
		switch tok.(type) {
		case json.Delim:
		default:
			j.replaceState(emiteObjectValueState)
			j.stagedToken = tok
			return JsonElement{local: j.currentName()}, false, nil
		}
	}

	if j.currentState() == readArrayState {
		switch tok.(type) {
		case json.Delim:
		default:
			j.replaceState(readArrayValueState)
			j.stagedToken = tok
			return JsonElement{local: j.currentName()}, false, nil
		}
	}

	switch t := tok.(type) {
	case json.Delim:
		return parseJsonDelim(j, t)
	case string:
		if j.currentState() == readObjectFieldState {
			j.replaceState(readObjectValueState)
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
		j.appendState(readObjectFieldState)

		if state == readObjectValueState || state == readArrayState {
			return JsonElement{local: j.currentName()}, false, nil
		}
	case "}":
		j.popState()

		if j.currentState() == readObjectValueState {
			j.replaceState(readObjectFieldState)
			return nil, true, nil
		}

		if j.currentState() == readArrayState {
			return nil, true, nil
		}
	case "[":
		j.appendState(readArrayState)
	case "]":
		j.popState()

		if j.currentState() == readObjectValueState {
			j.popName()
			j.replaceState(readObjectFieldState)
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
