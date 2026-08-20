package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSystemDefGrid(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "system_def_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	systemDefPath := filepath.Join(tempDir, "system.def")
	content := `[Title Info]
title = "My Game"

[Select Info]
rows = 4
columns = 8
wrapping = 1
showemptyboxes = 1
cell.size = 27,27
`
	_ = os.WriteFile(systemDefPath, []byte(content), 0644)

	grid := ParseSystemDefGrid(systemDefPath)
	if grid.Rows != 4 {
		t.Errorf("Expected 4 rows, got %d", grid.Rows)
	}
	if grid.Columns != 8 {
		t.Errorf("Expected 8 columns, got %d", grid.Columns)
	}
	if !grid.Wrapping || !grid.ShowEmptyBoxes {
		t.Errorf("Expected wrapping and showemptyboxes to be true")
	}
}

func TestSelectDefRoundTrip(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "select_def_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	charsDir := filepath.Join(tempDir, "chars", "kfm")
	_ = os.MkdirAll(charsDir, 0755)
	_ = os.WriteFile(filepath.Join(charsDir, "kfm.def"), []byte("[Info]\nname = kfm\ndisplayname = Kung Fu Man\nauthor = Elecbyte"), 0644)

	selectDefPath := filepath.Join(tempDir, "data", "select.def")
	_ = os.MkdirAll(filepath.Dir(selectDefPath), 0755)

	initialContent := `[Characters]
kfm, stages/mountains.def, music=sound/kfm.mp3, order=1
randomselect
empty
; kfm (Disabled)

[ExtraStages]
stages/dojo.def
stages/bridge.def
`
	_ = os.WriteFile(selectDefPath, []byte(initialContent), 0644)

	roster, err := GetProjectRoster(tempDir)
	if err != nil {
		t.Fatalf("GetProjectRoster failed: %v", err)
	}

	if len(roster.Slots) != 4 {
		t.Fatalf("Expected 4 slots, got %d", len(roster.Slots))
	}

	if roster.Slots[0].Character != "kfm" || roster.Slots[0].Order != 1 || roster.Slots[0].HomeStage != "stages/mountains.def" {
		t.Errorf("Slot 0 mismatch: %+v", roster.Slots[0])
	}
	if roster.Slots[1].Type != "randomselect" {
		t.Errorf("Slot 1 mismatch: %+v", roster.Slots[1])
	}
	if roster.Slots[2].Type != "empty" {
		t.Errorf("Slot 2 mismatch: %+v", roster.Slots[2])
	}
	if roster.Slots[3].Type != "disabled" {
		t.Errorf("Slot 3 mismatch: %+v", roster.Slots[3])
	}

	if len(roster.ExtraStages) != 2 {
		t.Errorf("Expected 2 extra stages, got %d", len(roster.ExtraStages))
	}

	// Test saving
	err = SaveProjectRoster(tempDir, *roster)
	if err != nil {
		t.Fatalf("SaveProjectRoster failed: %v", err)
	}

	// Re-parse
	roster2, err := GetProjectRoster(tempDir)
	if err != nil {
		t.Fatalf("Re-parsing failed: %v", err)
	}
	if len(roster2.Slots) != 4 {
		t.Fatalf("Expected 4 slots on re-read, got %d", len(roster2.Slots))
	}
}

func TestSelectDefTutorialSkipping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "select_tutorial_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	for _, name := range []string{"kfm", "kfm_zss", "kfm720", "cammy", "karin"} {
		charsDir := filepath.Join(tempDir, "chars", name)
		_ = os.MkdirAll(charsDir, 0755)
		_ = os.WriteFile(filepath.Join(charsDir, name+".def"), []byte("[Info]\nname = "+name+"\ndisplayname = "+name), 0644)
	}

	selectDefPath := filepath.Join(tempDir, "data", "select.def")
	_ = os.MkdirAll(filepath.Dir(selectDefPath), 0755)

	content := `[Characters]
 ;How to add characters
 ;Use the format:
 ; kfm, stages/mybg.def
 ; kfm/alt-kfm.def, stages/mybg.def
 ; kfm, random
 ;Insert your characters below.
;kfm_zss, stages/kfm.def
;kfm720, stages/kfm.def
cammy
karin
randomselect
`
	_ = os.WriteFile(selectDefPath, []byte(content), 0644)

	roster, err := GetProjectRoster(tempDir)
	if err != nil {
		t.Fatalf("GetProjectRoster failed: %v", err)
	}

	// Should only parse the 2 commented kfm variants under "Insert your characters below" + 2 active characters + 1 randomselect = 5 slots
	if len(roster.Slots) != 5 {
		t.Fatalf("Expected 5 slots (skipping tutorial kfm examples), got %d", len(roster.Slots))
	}

	if roster.Slots[0].Character != "kfm_zss" || roster.Slots[0].Type != "disabled" {
		t.Errorf("Expected slot 0 to be disabled kfm_zss, got %+v", roster.Slots[0])
	}
	if roster.Slots[1].Character != "kfm720" || roster.Slots[1].Type != "disabled" {
		t.Errorf("Expected slot 1 to be disabled kfm720, got %+v", roster.Slots[1])
	}
	if roster.Slots[2].Character != "cammy" || roster.Slots[2].Type != "character" {
		t.Errorf("Expected slot 2 to be cammy, got %+v", roster.Slots[2])
	}
}

