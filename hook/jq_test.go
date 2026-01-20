package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJQ(t *testing.T) {
	require := require.New(t)
	create := `{"action":"create","data":{"title":"Title Parsed","description":"Body Parsed"}}`
	comment := `{"action":"comment","data":{"issue":{"title":"Title Parsed"},"body":"Body Parsed"}}`
	for _, js := range []string{create, comment} {
		var payload any
		err := json.Unmarshal([]byte(js), &payload)
		require.Nil(err)
		path := ".data.title//.data.issue.title"
		title := getString(payload, path)
		require.Equal("Title Parsed", title, js)
		path = ".data.description//.data.body"
		body := getString(payload, path)
		require.Equal("Body Parsed", body, js)
	}
}
