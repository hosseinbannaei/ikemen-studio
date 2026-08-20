package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectStagesManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ikemen_stage_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.MkdirAll(filepath.Join(tempDir, "data"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "stages"), 0755)

	// Create sample select.def
	selectContent := `[Characters]
kfm, stages/stage0.def

[ExtraStages]
stages/stage0.def
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "select.def"), []byte(selectContent), 0644)

	// Create sample stage
	stageContent := `[Info]
name = "Test Stage"
displayname = "Test Arena"
author = "Tester"
mugenversion = "1.0"

[StageInfo]
zoffset = 220
xscale = 1.0
yscale = 1.0

[BGdef]
spr = stages/test.sff
bgmusic = sound/test_bgm.mp3
`
	_ = os.WriteFile(filepath.Join(tempDir, "stages", "test_stage.def"), []byte(stageContent), 0644)

	stages, err := GetProjectStages(tempDir)
	if err != nil {
		t.Fatalf("GetProjectStages failed: %v", err)
	}
	if len(stages) != 1 {
		t.Fatalf("Expected 1 stage, got %d", len(stages))
	}
	if stages[0].DisplayName != "Test Arena" {
		t.Errorf("Expected stage name 'Test Arena', got '%s'", stages[0].DisplayName)
	}
	if stages[0].Author != "Tester" {
		t.Errorf("Expected author 'Tester', got '%s'", stages[0].Author)
	}

	// Toggle Extra Stage ON
	err = ToggleStageExtraStage(tempDir, "stages/test_stage.def", true)
	if err != nil {
		t.Fatalf("ToggleStageExtraStage enable failed: %v", err)
	}

	stagesAfter, _ := GetProjectStages(tempDir)
	if !stagesAfter[0].IsExtraStage {
		t.Errorf("Expected stage to be marked as ExtraStage")
	}

	// Assign homestage to kfm
	err = SetFighterHomeStage(tempDir, "kfm", "stages/test_stage.def")
	if err != nil {
		t.Fatalf("SetFighterHomeStage failed: %v", err)
	}

	stagesCharAssigned, _ := GetProjectStages(tempDir)
	if len(stagesCharAssigned[0].AssignedCharacters) == 0 || stagesCharAssigned[0].AssignedCharacters[0] != "kfm" {
		t.Errorf("Expected assigned character 'kfm', got %v", stagesCharAssigned[0].AssignedCharacters)
	}

	// Toggle Extra Stage OFF
	err = ToggleStageExtraStage(tempDir, "stages/test_stage.def", false)
	if err != nil {
		t.Fatalf("ToggleStageExtraStage disable failed: %v", err)
	}
	stagesDisabled, _ := GetProjectStages(tempDir)
	if stagesDisabled[0].IsExtraStage {
		t.Errorf("Expected stage to NOT be an ExtraStage")
	}
}

func TestProjectMotifsManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ikemen_motif_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.MkdirAll(filepath.Join(tempDir, "data", "custom_motif"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "save"), 0755)

	systemContent := `[Info]
name = "Custom Motif"
author = "MotifDev"

[Select Info]
columns = 10
rows = 6

[Title Info]
localcoord = 1920, 1080
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "custom_motif", "system.def"), []byte(systemContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "save", "config.ini"), []byte("[Options]\nMotif = data/custom_motif/system.def\n"), 0644)

	motifs, err := GetProjectMotifs(tempDir)
	if err != nil {
		t.Fatalf("GetProjectMotifs failed: %v", err)
	}
	if len(motifs) != 1 {
		t.Fatalf("Expected 1 motif, got %d", len(motifs))
	}
	if !motifs[0].IsActive {
		t.Errorf("Expected motif to be active")
	}
	if motifs[0].GridColumns != 10 || motifs[0].GridRows != 6 {
		t.Errorf("Expected grid 10x6, got %dx%d", motifs[0].GridColumns, motifs[0].GridRows)
	}
	if motifs[0].Resolution != "1920x1080" {
		t.Errorf("Expected resolution 1920x1080, got %s", motifs[0].Resolution)
	}
}

func TestProjectLifebarsManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ikemen_lifebar_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.MkdirAll(filepath.Join(tempDir, "data", "lifebars", "mvc"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "data", "system.def"), []byte("[Files]\nfight = data/lifebars/mvc/fight.def\n"), 0644)

	fightContent := `[Info]
name = "MvC Style Lifebars"
author = "LifebarCreator"

[Files]
sff = fight.sff
snd = fight.snd
font1 = font/f-4x7.def
font2 = font/f-6x9.def
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "lifebars", "mvc", "fight.def"), []byte(fightContent), 0644)

	lifebars, err := GetProjectLifebars(tempDir)
	if err != nil {
		t.Fatalf("GetProjectLifebars failed: %v", err)
	}
	if len(lifebars) != 1 {
		t.Fatalf("Expected 1 lifebar, got %d", len(lifebars))
	}
	if !lifebars[0].IsActive {
		t.Errorf("Expected lifebar to be active")
	}
	if lifebars[0].FontCount != 2 {
		t.Errorf("Expected 2 fonts, got %d", lifebars[0].FontCount)
	}

	// Switch active lifebar
	_ = os.MkdirAll(filepath.Join(tempDir, "data", "lifebars", "sf3"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "data", "lifebars", "sf3", "fight.def"), []byte("[Info]\nname = SF3\n"), 0644)

	err = SetActiveLifebar(tempDir, "data/lifebars/sf3/fight.def")
	if err != nil {
		t.Fatalf("SetActiveLifebar failed: %v", err)
	}

	lifebarsUpdated, _ := GetProjectLifebars(tempDir)
	for _, lb := range lifebarsUpdated {
		if strings.Contains(lb.Key, "sf3") && !lb.IsActive {
			t.Errorf("Expected SF3 lifebar to be active now")
		}
	}
}

func TestProjectAudioManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ikemen_audio_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.MkdirAll(filepath.Join(tempDir, "data"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "sound"), 0755)

	systemContent := `[Music]
title.bgm = sound/title_theme.mp3
select.bgm = sound/select_theme.ogg
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "system.def"), []byte(systemContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "sound", "title_theme.mp3"), []byte("sample_audio_data"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "sound", "select_theme.ogg"), []byte("sample_audio_data"), 0644)

	audioList, err := GetProjectAudio(tempDir)
	if err != nil {
		t.Fatalf("GetProjectAudio failed: %v", err)
	}
	if len(audioList) != 2 {
		t.Fatalf("Expected 2 audio files, got %d", len(audioList))
	}

	for _, a := range audioList {
		if a.FileName == "title_theme.mp3" {
			if len(a.AssignedEvents) == 0 || a.AssignedEvents[0] != "Title Screen" {
				t.Errorf("Expected 'Title Screen' assignment, got %v", a.AssignedEvents)
			}
		}
	}

	// Update victory BGM
	err = SetSystemBGM(tempDir, "victory", "sound/title_theme.mp3")
	if err != nil {
		t.Fatalf("SetSystemBGM failed: %v", err)
	}

	audioUpdated, _ := GetProjectAudio(tempDir)
	for _, a := range audioUpdated {
		if a.FileName == "title_theme.mp3" {
			if len(a.AssignedEvents) < 2 {
				t.Errorf("Expected multiple assignments for title_theme.mp3, got %v", a.AssignedEvents)
			}
		}
	}
}

func TestProjectFontsAndStoryboards(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ikemen_fonts_story_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	_ = os.MkdirAll(filepath.Join(tempDir, "data"), 0755)
	_ = os.MkdirAll(filepath.Join(tempDir, "font"), 0755)

	systemContent := `[Files]
font1 = font/f-4x7.def
intro.storyboard = data/intro.def
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "system.def"), []byte(systemContent), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "font", "f-4x7.def"), []byte("[Font]\n"), 0644)

	introContent := `[SceneDef]
spr = intro.sff

[Scene 0]
fadein.time = 30
bgm = sound/intro.mp3

[Scene 1]
fadein.time = 20
`
	_ = os.WriteFile(filepath.Join(tempDir, "data", "intro.def"), []byte(introContent), 0644)

	fonts, err := GetProjectFonts(tempDir)
	if err != nil {
		t.Fatalf("GetProjectFonts failed: %v", err)
	}
	if len(fonts) != 1 {
		t.Fatalf("Expected 1 font, got %d", len(fonts))
	}
	if len(fonts[0].SystemSlotMappings) == 0 {
		t.Errorf("Expected font slot mapping for f-4x7.def")
	}

	storyboards, err := GetProjectStoryboards(tempDir)
	if err != nil {
		t.Fatalf("GetProjectStoryboards failed: %v", err)
	}
	if len(storyboards) != 1 {
		t.Fatalf("Expected 1 storyboard, got %d", len(storyboards))
	}
	if storyboards[0].SceneCount != 2 {
		t.Errorf("Expected 2 scenes in intro, got %d", storyboards[0].SceneCount)
	}
	if len(storyboards[0].AssignedSlots) == 0 || storyboards[0].AssignedSlots[0] != "Opening Intro" {
		t.Errorf("Expected 'Opening Intro' slot assignment, got %v", storyboards[0].AssignedSlots)
	}
}
