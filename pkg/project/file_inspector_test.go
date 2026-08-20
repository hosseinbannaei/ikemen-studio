package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectProjectFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "inspect_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// 1. Create a character DEF file
	defContent := `[Info]
name = "Ryu"
displayname = "Ryu Hoshi"
author = "Capcom"

[Files]
cns = ryu.cns
cmd = ryu.cmd
anim = ryu.air
`
	_ = os.WriteFile(filepath.Join(tempDir, "ryu.def"), []byte(defContent), 0644)

	resDef, err := InspectProjectFile(tempDir, "ryu.def")
	if err != nil {
		t.Fatalf("InspectProjectFile for ryu.def failed: %v", err)
	}
	if resDef.Category != "fighter" || resDef.DisplayName != "Ryu Hoshi" {
		t.Errorf("Unexpected def metadata: %+v", resDef)
	}
	if len(resDef.Sections) != 2 {
		t.Errorf("Expected 2 sections in def, got %d", len(resDef.Sections))
	}

	// 2. Create an AIR file
	airContent := `; Standing animation
[Begin Action 0] ; Idle stance
Clsn2: 1
  Clsn2[0] = -10, -80, 10, 0
0,0, 0,0, 6
0,1, 0,0, 6
Loopstart
0,2, 0,0, 6

[Begin Action 200] ; Light punch
Clsn2: 1
  Clsn2[0] = -10, -80, 10, 0
Clsn1: 1
  Clsn1[0] = 5, -60, 25, -45
200,0, 0,0, 3
200,1, 0,0, 4
`
	_ = os.WriteFile(filepath.Join(tempDir, "ryu.air"), []byte(airContent), 0644)

	resAir, err := InspectProjectFile(tempDir, "ryu.air")
	if err != nil {
		t.Fatalf("InspectProjectFile for ryu.air failed: %v", err)
	}
	if len(resAir.AnimActions) != 2 {
		t.Fatalf("Expected 2 anim actions, got %d", len(resAir.AnimActions))
	}
	if !resAir.AnimActions[0].HasLoop || !resAir.AnimActions[1].HasHitbox {
		t.Errorf("Action flags mismatch: %+v", resAir.AnimActions)
	}

	// 3. Create a CMD file
	cmdContent := `[Command]
name = "hadouken"
command = ~D, DF, F, a
time = 15
buffer.time = 5

[Command]
name = "shoryuken"
command = ~F, D, DF, a
time = 20
`
	_ = os.WriteFile(filepath.Join(tempDir, "ryu.cmd"), []byte(cmdContent), 0644)

	resCmd, err := InspectProjectFile(tempDir, "ryu.cmd")
	if err != nil {
		t.Fatalf("InspectProjectFile for ryu.cmd failed: %v", err)
	}
	if len(resCmd.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(resCmd.Commands))
	}
	if resCmd.Commands[0].Name != "hadouken" || resCmd.Commands[0].BufferTime != 5 {
		t.Errorf("Command metadata mismatch: %+v", resCmd.Commands[0])
	}

	// 4. Create a CNS file
	cnsContent := `[Data]
life = 1000
attack = 100
defence = 100

[Size]
xscale = 1
yscale = 1

[StateDef 200, Light Punch]
type = S
movetype = A
physics = S
anim = 200

[State 200, 1]
type = HitDef
trigger1 = time = 0
damage = 25, 0
`
	_ = os.WriteFile(filepath.Join(tempDir, "ryu.cns"), []byte(cnsContent), 0644)

	resCns, err := InspectProjectFile(tempDir, "ryu.cns")
	if err != nil {
		t.Fatalf("InspectProjectFile for ryu.cns failed: %v", err)
	}
	if len(resCns.StateDefs) != 1 || resCns.StateDefs[0].StateNo != 200 {
		t.Fatalf("Expected StateDef 200, got: %+v", resCns.StateDefs)
	}

	// 5. Test safe Read and Save Project File
	err = SaveProjectFile(tempDir, "notes/readme.txt", "Hello Ikemen Studio")
	if err != nil {
		t.Fatalf("SaveProjectFile failed: %v", err)
	}

	readContent, err := ReadProjectFile(tempDir, "notes/readme.txt")
	if err != nil || readContent != "Hello Ikemen Studio" {
		t.Fatalf("ReadProjectFile failed, content: %s, err: %v", readContent, err)
	}

	// 6. Test path traversal rejection
	_, err = ReadProjectFile(tempDir, "../../etc/passwd")
	if err == nil {
		t.Errorf("Expected path traversal error, got nil")
	}
}
