package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// FileInspectionResult provides rich structural analysis of any editable project file.
type FileInspectionResult struct {
	RelPath      string                 `json:"rel_path"`
	FileType     string                 `json:"file_type"` // "def", "cns", "zss", "cmd", "air", "ini", "lua", "txt", "shader"
	Category     string                 `json:"category"`  // "fighter", "stage", "motif", "lifebar", "storyboard", "font", "config", "script", "animation", "commands"
	DisplayName  string                 `json:"display_name"`
	TotalLines   int                    `json:"total_lines"`
	SizeBytes    int64                  `json:"size_bytes"`
	Sections     []FileSectionSummary   `json:"sections"`
	KeyValues    map[string]string      `json:"key_values"`
	AnimActions  []AnimActionSummary    `json:"anim_actions,omitempty"`
	Commands     []CommandEntrySummary  `json:"commands,omitempty"`
	StateDefs    []StateDefSummary      `json:"state_defs,omitempty"`
	RawContent   string                 `json:"raw_content"`
	IsEditable   bool                   `json:"is_editable"`
	SyntaxMode   string                 `json:"syntax_mode"` // "ini", "zss", "lua", "glsl", "plain"
}

// FileSectionSummary represents an INI or DEF section bracket [SectionName].
type FileSectionSummary struct {
	Name      string `json:"name"`
	LineStart int    `json:"line_start"`
	LineEnd   int    `json:"line_end"`
	ItemCount int    `json:"item_count"`
}

// AnimActionSummary represents an animation action [Begin Action <n>] in an .air file.
type AnimActionSummary struct {
	ActionNo    int    `json:"action_no"`
	Description string `json:"description,omitempty"`
	FrameCount  int    `json:"frame_count"`
	TotalTicks  int    `json:"total_ticks"`
	HasLoop     bool   `json:"has_loop"`
	HasHitbox   bool   `json:"has_hitbox"`
	HasHurtbox  bool   `json:"has_hurtbox"`
}

// CommandEntrySummary represents a registered command sequence in a .cmd file.
type CommandEntrySummary struct {
	Name       string `json:"name"`
	Command    string `json:"command"`
	Time       int    `json:"time"`
	BufferTime int    `json:"buffer_time"`
}

// StateDefSummary represents a state definition [StateDef <n>] in a .cns or .zss file.
type StateDefSummary struct {
	StateNo         int    `json:"state_no"`
	Name            string `json:"name,omitempty"`
	Type            string `json:"type"`      // S (Stand), C (Crouch), A (Air), L (Liedown)
	MoveType        string `json:"move_type"` // A (Attack), I (Idle), H (Hit)
	Physics         string `json:"physics"`   // S, C, A, N (None)
	Anim            int    `json:"anim"`
	ControllerCount int    `json:"controller_count"`
}

// ReadProjectFile safely reads text content of a file within a project.
func ReadProjectFile(projectDir, relPath string) (string, error) {
	cleanPath, err := sanitizeProjectRelPath(projectDir, relPath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", relPath, err)
	}

	return string(data), nil
}

// SaveProjectFile safely writes text content to a file within a project.
func SaveProjectFile(projectDir, relPath, content string) error {
	cleanPath, err := sanitizeProjectRelPath(projectDir, relPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", relPath, err)
	}

	return os.WriteFile(cleanPath, []byte(content), 0644)
}

// InspectProjectFile analyzes any project file and extracts metadata, sections, and structural metrics.
func InspectProjectFile(projectDir, relPath string) (*FileInspectionResult, error) {
	cleanPath, err := sanitizeProjectRelPath(projectDir, relPath)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("file not found %s: %w", relPath, err)
	}

	contentBytes, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", relPath, err)
	}

	content := string(contentBytes)
	ext := strings.ToLower(filepath.Ext(cleanPath))
	baseName := filepath.Base(cleanPath)

	result := &FileInspectionResult{
		RelPath:     filepath.ToSlash(relPath),
		DisplayName: baseName,
		TotalLines:  strings.Count(content, "\n") + 1,
		SizeBytes:   fi.Size(),
		Sections:    make([]FileSectionSummary, 0),
		KeyValues:   make(map[string]string),
		RawContent:  content,
		IsEditable:  true,
		SyntaxMode:  "plain",
	}

	switch ext {
	case ".def":
		result.FileType = "def"
		result.SyntaxMode = "ini"
		inspectDefFile(content, result)
	case ".cns":
		result.FileType = "cns"
		result.Category = "fighter"
		result.SyntaxMode = "ini"
		inspectCnsFile(content, result)
	case ".zss":
		result.FileType = "zss"
		result.Category = "script"
		result.SyntaxMode = "zss"
		inspectZssFile(content, result)
	case ".cmd":
		result.FileType = "cmd"
		result.Category = "commands"
		result.SyntaxMode = "ini"
		inspectCmdFile(content, result)
	case ".air":
		result.FileType = "air"
		result.Category = "animation"
		result.SyntaxMode = "ini"
		inspectAirFile(content, result)
	case ".ini":
		result.FileType = "ini"
		result.Category = "config"
		result.SyntaxMode = "ini"
		inspectIniFile(content, result)
	case ".lua":
		result.FileType = "lua"
		result.Category = "script"
		result.SyntaxMode = "lua"
	case ".frag", ".vert", ".glsl":
		result.FileType = "shader"
		result.Category = "shader"
		result.SyntaxMode = "glsl"
	default:
		result.FileType = strings.TrimPrefix(ext, ".")
		result.Category = "text"
		result.SyntaxMode = "plain"
	}

	return result, nil
}

func inspectDefFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	var curSection *FileSectionSummary

	sectionRegex := regexp.MustCompile(`^\s*\[\s*([^\]]+?)\s*\]`)
	kvRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_.\-]+)\s*=\s*(.*?)(?:;.*)?$`)

	lowerContent := strings.ToLower(content)
	if strings.Contains(lowerContent, "[stagedef]") || strings.Contains(lowerContent, "[camera]") || strings.Contains(lowerContent, "[bgdef]") {
		res.Category = "stage"
	} else if strings.Contains(lowerContent, "[title info]") || strings.Contains(lowerContent, "[select info]") {
		res.Category = "motif"
	} else if strings.Contains(lowerContent, "[lifebar]") || strings.Contains(lowerContent, "[fightfx]") {
		res.Category = "lifebar"
	} else if strings.Contains(lowerContent, "[scenedef]") || strings.Contains(lowerContent, "[scene 0]") {
		res.Category = "storyboard"
	} else if strings.Contains(lowerContent, "[fnt v2]") || strings.Contains(lowerContent, "type = truetype") || strings.Contains(lowerContent, "type = bitmap") {
		res.Category = "font"
	} else {
		res.Category = "fighter"
	}

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := sectionRegex.FindStringSubmatch(line); len(sm) > 1 {
			secName := strings.TrimSpace(sm[1])
			if curSection != nil {
				curSection.LineEnd = lineNo - 1
				res.Sections = append(res.Sections, *curSection)
			}
			curSection = &FileSectionSummary{
				Name:      secName,
				LineStart: lineNo,
				LineEnd:   lineNo,
				ItemCount: 0,
			}
			continue
		}

		if curSection != nil {
			curSection.LineEnd = lineNo
		}

		if kv := kvRegex.FindStringSubmatch(line); len(kv) > 2 {
			k := strings.ToLower(strings.TrimSpace(kv[1]))
			v := strings.Trim(strings.TrimSpace(kv[2]), `"`)
			res.KeyValues[k] = v
			if curSection != nil {
				curSection.ItemCount++
			}
			if k == "displayname" && v != "" {
				res.DisplayName = v
			} else if k == "name" && v != "" && res.DisplayName == filepath.Base(res.RelPath) {
				res.DisplayName = v
			}

		}
	}

	if curSection != nil {
		curSection.LineEnd = lineNo
		res.Sections = append(res.Sections, *curSection)
	}
}

func inspectCnsFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	var curSection *FileSectionSummary
	var curState *StateDefSummary

	stateDefRegex := regexp.MustCompile(`(?i)^\s*\[\s*StateDef\s+(-?\d+)(?:,\s*([^\]]*))?\s*\]`)
	stateCtrlRegex := regexp.MustCompile(`(?i)^\s*\[\s*State\s+([^\]]+)\s*\]`)
	sectionRegex := regexp.MustCompile(`^\s*\[\s*([^\]]+?)\s*\]`)
	kvRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_.\-]+)\s*=\s*(.*?)(?:;.*)?$`)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := stateDefRegex.FindStringSubmatch(line); len(sm) > 1 {
			num, _ := strconv.Atoi(sm[1])
			name := ""
			if len(sm) > 2 {
				name = strings.TrimSpace(sm[2])
			}
			if curState != nil {
				res.StateDefs = append(res.StateDefs, *curState)
			}
			curState = &StateDefSummary{
				StateNo:  num,
				Name:     name,
				Type:     "S",
				MoveType: "I",
				Physics:  "S",
			}
			if curSection != nil {
				curSection.LineEnd = lineNo - 1
				res.Sections = append(res.Sections, *curSection)
			}
			curSection = &FileSectionSummary{
				Name:      fmt.Sprintf("StateDef %d", num),
				LineStart: lineNo,
				LineEnd:   lineNo,
			}
			continue
		}

		if sc := stateCtrlRegex.FindStringSubmatch(line); len(sc) > 1 {
			if curState != nil {
				curState.ControllerCount++
			}
		}

		if sm := sectionRegex.FindStringSubmatch(line); len(sm) > 1 && !strings.HasPrefix(strings.ToLower(sm[1]), "state") {
			secName := strings.TrimSpace(sm[1])
			if curSection != nil {
				curSection.LineEnd = lineNo - 1
				res.Sections = append(res.Sections, *curSection)
			}
			curSection = &FileSectionSummary{
				Name:      secName,
				LineStart: lineNo,
				LineEnd:   lineNo,
			}
		}

		if curSection != nil {
			curSection.LineEnd = lineNo
		}

		if kv := kvRegex.FindStringSubmatch(line); len(kv) > 2 {
			k := strings.ToLower(strings.TrimSpace(kv[1]))
			v := strings.Trim(strings.TrimSpace(kv[2]), `"`)
			res.KeyValues[k] = v
			if curState != nil {
				switch k {
				case "type":
					curState.Type = strings.ToUpper(v)
				case "movetype":
					curState.MoveType = strings.ToUpper(v)
				case "physics":
					curState.Physics = strings.ToUpper(v)
				case "anim":
					if aNum, err := strconv.Atoi(v); err == nil {
						curState.Anim = aNum
					}
				}
			}
		}
	}

	if curState != nil {
		res.StateDefs = append(res.StateDefs, *curState)
	}
	if curSection != nil {
		curSection.LineEnd = lineNo
		res.Sections = append(res.Sections, *curSection)
	}
}

func inspectZssFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	stateDefRegex := regexp.MustCompile(`(?i)^\s*\[\s*StateDef\s+(-?\d+)(?:,\s*([^\]]*))?\s*\]`)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := stateDefRegex.FindStringSubmatch(line); len(sm) > 1 {
			num, _ := strconv.Atoi(sm[1])
			name := ""
			if len(sm) > 2 {
				name = strings.TrimSpace(sm[2])
			}
			res.StateDefs = append(res.StateDefs, StateDefSummary{
				StateNo: num,
				Name:    name,
			})
			res.Sections = append(res.Sections, FileSectionSummary{
				Name:      fmt.Sprintf("StateDef %d", num),
				LineStart: lineNo,
			})
		}
	}
}

func inspectCmdFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	var curCmd *CommandEntrySummary
	inCommandBlock := false

	sectionRegex := regexp.MustCompile(`^\s*\[\s*([^\]]+?)\s*\]`)
	kvRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_.\-]+)\s*=\s*(.*?)(?:;.*)?$`)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := sectionRegex.FindStringSubmatch(line); len(sm) > 1 {
			secName := strings.TrimSpace(sm[1])
			if strings.EqualFold(secName, "command") {
				inCommandBlock = true
				if curCmd != nil && curCmd.Name != "" {
					res.Commands = append(res.Commands, *curCmd)
				}
				curCmd = &CommandEntrySummary{Time: 15, BufferTime: 1}
			} else {
				inCommandBlock = false
				if curCmd != nil && curCmd.Name != "" {
					res.Commands = append(res.Commands, *curCmd)
					curCmd = nil
				}
			}
			continue
		}

		if inCommandBlock && curCmd != nil {
			if kv := kvRegex.FindStringSubmatch(line); len(kv) > 2 {
				k := strings.ToLower(strings.TrimSpace(kv[1]))
				v := strings.Trim(strings.TrimSpace(kv[2]), `"`)
				switch k {
				case "name":
					curCmd.Name = v
				case "command":
					curCmd.Command = v
				case "time":
					if t, err := strconv.Atoi(v); err == nil {
						curCmd.Time = t
					}
				case "buffer.time":
					if bt, err := strconv.Atoi(v); err == nil {
						curCmd.BufferTime = bt
					}
				}
			}
		}
	}

	if curCmd != nil && curCmd.Name != "" {
		res.Commands = append(res.Commands, *curCmd)
	}
}

func inspectAirFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	var curAction *AnimActionSummary

	actionRegex := regexp.MustCompile(`(?i)^\s*\[\s*Begin\s+Action\s+(\d+)\s*\](?:;*(.*))?`)
	clsnRegex := regexp.MustCompile(`(?i)^\s*Clsn([12])(?::|\[)`)
	loopRegex := regexp.MustCompile(`(?i)^\s*Loopstart`)
	frameRegex := regexp.MustCompile(`^\s*-?\d+\s*,\s*-?\d+\s*,\s*-?\d+\s*,\s*-?\d+\s*,\s*(-?\d+)`)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := actionRegex.FindStringSubmatch(line); len(sm) > 1 {
			num, _ := strconv.Atoi(sm[1])
			desc := ""
			if len(sm) > 2 {
				desc = strings.TrimSpace(sm[2])
			}
			if curAction != nil {
				res.AnimActions = append(res.AnimActions, *curAction)
			}
			curAction = &AnimActionSummary{
				ActionNo:    num,
				Description: desc,
			}
			res.Sections = append(res.Sections, FileSectionSummary{
				Name:      fmt.Sprintf("Begin Action %d", num),
				LineStart: lineNo,
			})
			continue
		}

		if curAction != nil {
			if clsn := clsnRegex.FindStringSubmatch(line); len(clsn) > 1 {
				if clsn[1] == "1" {
					curAction.HasHitbox = true
				} else if clsn[1] == "2" {
					curAction.HasHurtbox = true
				}
			} else if loopRegex.MatchString(line) {
				curAction.HasLoop = true
			} else if fm := frameRegex.FindStringSubmatch(line); len(fm) > 1 {
				curAction.FrameCount++
				ticks, _ := strconv.Atoi(fm[1])
				if ticks > 0 {
					curAction.TotalTicks += ticks
				}
			}
		}
	}

	if curAction != nil {
		res.AnimActions = append(res.AnimActions, *curAction)
	}
}

func inspectIniFile(content string, res *FileInspectionResult) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	var curSection *FileSectionSummary
	sectionRegex := regexp.MustCompile(`^\s*\[\s*([^\]]+?)\s*\]`)
	kvRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_.\-]+)\s*=\s*(.*?)(?:;.*)?$`)

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		if sm := sectionRegex.FindStringSubmatch(line); len(sm) > 1 {
			secName := strings.TrimSpace(sm[1])
			if curSection != nil {
				curSection.LineEnd = lineNo - 1
				res.Sections = append(res.Sections, *curSection)
			}
			curSection = &FileSectionSummary{
				Name:      secName,
				LineStart: lineNo,
				LineEnd:   lineNo,
			}
			continue
		}

		if curSection != nil {
			curSection.LineEnd = lineNo
		}

		if kv := kvRegex.FindStringSubmatch(line); len(kv) > 2 {
			k := strings.ToLower(strings.TrimSpace(kv[1]))
			v := strings.Trim(strings.TrimSpace(kv[2]), `"`)
			res.KeyValues[k] = v
			if curSection != nil {
				curSection.ItemCount++
			}
		}
	}

	if curSection != nil {
		curSection.LineEnd = lineNo
		res.Sections = append(res.Sections, *curSection)
	}
}

// sanitizeProjectRelPath ensures that a relative path stays within the project boundary.
func sanitizeProjectRelPath(projectDir, relPath string) (string, error) {
	cleanRel := filepath.Clean(filepath.ToSlash(strings.TrimSpace(relPath)))
	if strings.HasPrefix(cleanRel, "../") || cleanRel == ".." {
		return "", fmt.Errorf("path traversal attempt detected: %s", relPath)
	}

	fullPath := filepath.Join(projectDir, cleanRel)
	cleanFull := filepath.Clean(fullPath)
	cleanProj := filepath.Clean(projectDir)

	if !strings.HasPrefix(cleanFull, cleanProj) {
		return "", fmt.Errorf("resolved path %s is outside project directory %s", cleanFull, cleanProj)
	}

	return cleanFull, nil
}
