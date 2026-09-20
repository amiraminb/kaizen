package cmd

import (
	"bytes"
	"testing"
)

func TestUrgeCommandPrintsReflectionQuestions(t *testing.T) {
	var output bytes.Buffer
	cmd := urgeCmd
	cmd.SetOut(&output)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("urge command returned error: %v", err)
	}

	want := `Before giving in to an urge, pause and answer:
1. What is this thought about?
2. What are the feelings I have about this thought?
3. What are the physical feelings I have?
4. What are the thoughts about the main thought?
5. What do you want to do about this?

Take a note of your answers with: kaizen note <habit>
`
	if output.String() != want {
		t.Errorf("urge output = %q, want %q", output.String(), want)
	}
}
