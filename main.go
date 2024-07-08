package main

import (
	. "fmt"
	"os"
	. "strconv"
	. "strings"
	"syscall"
	"unsafe"
)

const (
	KeyBackspace = 8
	Tab          = 9
	KeyEnter     = 13
	KeyEscape    = 27

	VERSION = "0.1.5"
)

var (
	THEMES_FOLDER = os.ExpandEnv("$HOME/.config/scratchpad/themes")

	NERD_FONT = false
	UNICODE   = false

	LINE_NUM_FG      = "\x1b[38;5;239m"
	LINE_NUM_BG      = "\x1b[48;5;234m"
	TEXT_FG          = "\x1b[38;5;15m"
	TEXT_BG          = "\x1b[48;5;234m"
	EMPTY_LINE_FG    = "\x1b[38;5;236m"
	EMPTY_LINE_BG    = "\x1b[48;5;234m"
	STATUS_LINE_FG   = "\x1b[38;5;15m"
	STATUS_LINE_BG   = "\x1b[48;5;234m"
	SELECTED_NUM_FG  = "\x1b[38;5;11m"
	SELECTED_NUM_BG  = "\x1b[48;5;239m"
	SELECTED_TEXT_FG = "\x1b[38;5;11m"
	SELECTED_TEXT_BG = "\x1b[48;5;239m"

	H1 = "\x1b[38;5;14m"
	H2 = "\x1b[38;5;13m"
	H3 = "\x1b[38;5;12m"
	H4 = "\x1b[38;5;11m"
	H5 = "\x1b[38;5;10m"
	H6 = "\x1b[38;5;9m"

	LINENUM      = LINE_NUM_FG + LINE_NUM_BG
	LINETEXT     = TEXT_FG + TEXT_BG
	EMPTYLINE    = EMPTY_LINE_FG + EMPTY_LINE_BG
	STATUSLINE   = STATUS_LINE_FG + STATUS_LINE_BG
	SELECTEDNUM  = SELECTED_NUM_FG + SELECTED_NUM_BG
	SELECTEDTEXT = SELECTED_TEXT_FG + SELECTED_TEXT_BG
)

type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func restoreTerminal(oldState *syscall.Termios) {
	Print("\x1b[?25h")
	if oldState != nil {
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0); err != 0 {
			Println("Error restoring terminal:", err)
		}
	}
}

func setRawTerminal() (*syscall.Termios, error) {
	Print("\x1b[?25l")
	oldState := &syscall.Termios{}
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0); err != 0 {
		return nil, err
	}

	newState := *oldState

	newState.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	newState.Oflag &^= syscall.OPOST
	newState.Cflag |= syscall.CS8
	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	newState.Cc[syscall.VMIN] = 0
	newState.Cc[syscall.VTIME] = 1

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return oldState, nil
}

func readKey() (string, error) {
	var buf [3]byte
	for {
		n, err := os.Stdin.Read(buf[:])
		if err != nil {
			if err.Error() == "EOF" {
				return "", nil
			}
			return "", err
		}

		if n == 1 {
			switch buf[0] {
			case KeyBackspace, '\x7f':
				return "backspace", nil
			case KeyEscape:
				return "esc", nil
			case KeyEnter:
				return "enter", nil
			case Tab:
				return "tab", nil
			default:
				return string(buf[0]), nil
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			case 'C':
				return "right", nil
			case 'D':
				return "left", nil
			}
		}
	}
}

func isChar(key string) bool {
	if len(key) != 1 {
		return false
	}
	for _, c := range key {
		if c < ' ' || c > '~' {
			return false
		}
	}
	return true
}

func isCtrl(key string, c byte) bool {
	if len(key) != 1 {
		return false
	}
	return key[0] == c-'a'+1
}

func getSize(fd int) (*Winsize, error) {
	ws := &Winsize{}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return nil, Errorf("error getting terminal size: %v", err)
	}
	return ws, nil
}

func length(s string) int {
	return len([]rune(s))
}

func clearScreen() {
	Print("\x1b[2J")
	Print("\x1b[3J")
	Print("\x1b[H")
	os.Stdout.Sync()
}

func scratchPad(buffer string) error {
	var savePath string

	saving := false
	previewMode := false
	offset := 0
	pos2d := []int{0, 0}
	cursorPos := []int{1, 4}
	pos1d := 0

	for {
		var frame []string

		ws, err := getSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}

		intLines := 0
		lines := Split(buffer, "\n")
		numPadding := len(Sprint(len(lines) + offset))
		var lineWrap string
		if NERD_FONT {
			lineWrap = "󱞩"
		} else if UNICODE {
			lineWrap = "↪"
		} else {
			lineWrap = ">"
		}

		for i, line := range lines[offset:] {
			if intLines == int(ws.Row)-1 {
				break
			}
			var lineNum, lineText string
			if i+offset == pos2d[0] && !saving && !previewMode {
				lineNum = SELECTEDNUM
				lineText = SELECTEDTEXT
			} else {
				lineNum = LINENUM
				lineText = LINETEXT
			}

			var addSpaces int
			if previewMode {
				// isBold := false
				// newLine := ""
				// for k := 0; k < length(line); k++ {
				// 	if line[k] == '*' {
				// 		if length(line) > k+1 && line[k+1] == '*' {
				// 			if isBold {
				// 				addSpaces += len("\x1b[22m")
				// 				newLine += "\x1b[22m"
				// 				isBold = false
				// 			} else {
				// 				addSpaces += len("\x1b[1m")
				// 				newLine += "\x1b[1m"
				// 				isBold = true
				// 			}
				// 			k++
				// 		} else {
				// 			newLine += "*"
				// 		}
				// 	} else {
				// 		newLine += string(line[k])
				// 	}
				// }
				// line = newLine
				//
				// isItalic := false
				// newLine = ""
				// for k := 0; k < length(line); k++ {
				// 	if line[k] == '_' {
				// 		if isItalic {
				// 			addSpaces += len("\x1b[23m")
				// 			newLine += "\x1b[23m"
				// 			isItalic = false
				// 		} else {
				// 			addSpaces += len("\x1b[3m")
				// 			newLine += "\x1b[3m"
				// 			isItalic = true
				// 		}
				// 	} else {
				// 		newLine += string(line[k])
				// 	}
				// }
				// line = newLine

				if HasPrefix(line, "# ") {
					lineText = "\x1b[1m" + H1
					line = Replace(line, "# ", "", 1)
				} else if HasPrefix(line, "## ") {
					lineText = "\x1b[1m" + H2
					line = Replace(line, "## ", "", 1)
				} else if HasPrefix(line, "### ") {
					lineText = "\x1b[1m" + H3
					line = Replace(line, "### ", "", 1)
				} else if HasPrefix(line, "#### ") {
					lineText = "\x1b[1m" + H4
					line = Replace(line, "#### ", "", 1)
				} else if HasPrefix(line, "##### ") {
					lineText = "\x1b[1m" + H5
					line = Replace(line, "##### ", "", 1)
				} else if HasPrefix(line, "###### ") {
					lineText = "\x1b[1m" + H6
					line = Replace(line, "###### ", "", 1)
				} else {
					if HasPrefix(line, "> ") {
						newLine := ""
						for HasPrefix(line, "> ") {
							line = Replace(line, "> ", "", 1)
							if NERD_FONT {
								newLine += " "
							} else if UNICODE {
								newLine += "▏ "
							} else {
								newLine += "| "
							}
						}
						if HasPrefix(line, "- [x]") {
							if NERD_FONT {
								newLine += " "
							} else if UNICODE {
								newLine += "☒ "
							} else {
								newLine += "x "
							}
							line = Replace(line, "- [x]", "", 1)
						} else if HasPrefix(line, "- [ ]") {
							if NERD_FONT {
								newLine += "󰄱 "
							} else if UNICODE {
								newLine += "☐ "
							} else {
								newLine += "o "
							}
							line = Replace(line, "- [ ]", "", 1)
						} else if HasPrefix(line, "- ") {
							if NERD_FONT {
								newLine += " "
							} else if UNICODE {
								newLine += "• "
							} else {
								newLine += "- "
							}
							line = Replace(line, "- ", "", 1)
						}
						line = newLine + line
					} else {
						newLine := ""
						if HasPrefix(line, "- [x]") {
							if NERD_FONT {
								newLine += " "
							} else if UNICODE {
								newLine += "☒ "
							} else {
								newLine += "x "
							}
							line = Replace(line, "- [x]", "", 1)
						} else if HasPrefix(line, "- [ ]") {
							if NERD_FONT {
								newLine += "󰄱 "
							} else if UNICODE {
								newLine += "☐ "
							} else {
								newLine += "o "
							}
							line = Replace(line, "- [ ]", "", 1)
						} else if HasPrefix(line, "- ") {
							if NERD_FONT {
								newLine += " "
							} else if UNICODE {
								newLine += "• "
							} else {
								newLine += "- "
							}
							line = Replace(line, "- ", "", 1)
						}
						line = newLine + line
					}
				}
			}

			if length(line) > int(ws.Col)-numPadding-2 {
				splitLines := int(length(line) / (int(ws.Col) - numPadding - 2))
				if length(line)%(int(ws.Col)-numPadding-2) > 0 {
					splitLines++
				}
				shiftDown := 0
				for j := 0; j < splitLines; j++ {
					from := (int(ws.Col) - numPadding - 2) * j
					to := from + int(ws.Col) - numPadding - 2
					if to > length(line) {
						to = length(line)
					}
					if from <= pos2d[1] && pos2d[1] <= to {
						shiftDown = splitLines - j - 1
					}
					if j == 0 {
						frame = append(frame, Sprintf("%s%s%d %s %s%s\x1b[0m", lineNum, Repeat(" ", numPadding-len(Sprint(i+1+offset))), i+1+offset, lineText, line[from:to], Repeat(" ", int(ws.Col)-length(line[from:to])-numPadding-2)))
					} else {
						frame = append(frame, Sprintf("%s%s%s %s %s%s", lineNum, Repeat(" ", numPadding-1), lineWrap, lineText, line[from:to], Repeat(" ", int(ws.Col)-length(line[from:to])-numPadding-2)))
					}
					intLines++
				}
				if i+offset == pos2d[0] {
					cursorPos[0] = intLines - shiftDown
					cursorPos[1] = (pos2d[1] % (int(ws.Col) - numPadding - 2)) + numPadding + 3
					if pos2d[1]%(int(ws.Col)-numPadding-2) == 0 && pos2d[1]-(intLines-shiftDown-1)*(int(ws.Col)-numPadding-2) > 0 {
						cursorPos[0]++
						cursorPos[1] = numPadding + 3
						frame = append(frame, Sprintf("%s%s%s %s %s", lineNum, Repeat(" ", numPadding-1), lineWrap, lineText, Repeat(" ", int(ws.Col)-numPadding-2)))
						intLines++
					}
				}
			} else {
				frame = append(frame, Sprintf("%s%s%d %s %s%s\x1b[0m", lineNum, Repeat(" ", numPadding-len(Sprint(i+1+offset))), i+1+offset, lineText, line, Repeat(" ", int(ws.Col)-length(line)-numPadding-2+addSpaces)))
				intLines++
				if i+offset == pos2d[0] {
					cursorPos[0] = intLines
					cursorPos[1] = pos2d[1] + numPadding + 3
					if cursorPos[1] == int(ws.Col)+1 {
						cursorPos[0]++
						cursorPos[1] = numPadding + 3
						frame = append(frame, Sprintf("%s%s%s %s %s", lineNum, Repeat(" ", numPadding-1), lineWrap, lineText, Repeat(" ", int(ws.Col)-numPadding-2)))
						intLines++
					}
				}
			}
		}

		for i := 0; i < int(ws.Row)-intLines-1; i++ {
			frame = append(frame, Sprintf("%s~%s\x1b[0m", EMPTYLINE, Repeat(" ", int(ws.Col)-1)))
		}

		if saving {
			frame = append(frame, Sprintf("%s Save as: %s\x1b[48;5;252m \x1b[0m%s%s\x1b[0m", SELECTEDTEXT, savePath, SELECTEDTEXT, Repeat(" ", int(ws.Col)-11-len(savePath))))
		} else if previewMode {
			frame = append(frame, Sprintf("%s Preview Mode %s\x1b[0m", SELECTEDTEXT, Repeat(" ", int(ws.Col)-14)))
		} else {
			frame = append(frame, Sprintf("%s %d lines %s %d:%d \x1b[0m", STATUSLINE, len(lines), Repeat(" ", int(ws.Col)-11-len(Sprint(len(lines)))-len(Sprint(pos2d[0]+1))-len(Sprint(pos2d[1]+1))), pos2d[0]+1, pos2d[1]+1))
		}

		if previewMode || saving {
			clearScreen()
			Print(Join(frame, "\n\r"))
		} else {
			Print("\x1b[?25l")
			clearScreen()
			Print(Join(frame, "\n\r"))
			Printf("\x1b[%d;%dH", cursorPos[0], cursorPos[1]) // pos2d[0]-offset+1, pos2d[1]+numPadding+3
			Print("\x1b[?25h")
		}

		key, err := readKey()
		if err != nil {
			return err
		}
		if previewMode {
			if key == "esc" || isCtrl(key, 'p') {
				previewMode = false
				Print("\x1b[?25h")
			}
			continue
		} else {
			if key == "backspace" {
				if saving {
					if len(savePath) > 0 {
						savePath = savePath[:len(savePath)-1]
					}
				} else {
					if pos1d > 0 {
						lineLen := len(Split(buffer, "\n")[pos2d[0]])
						if pos2d[1] > 3 && HasSuffix(buffer[pos1d-4:pos1d], "    ") && pos2d[1]%4 == 0 {
							buffer = buffer[:pos1d-4] + buffer[pos1d:]
							pos1d -= 4
							pos2d[1] -= 4
						} else {
							buffer = buffer[:pos1d-1] + buffer[pos1d:]
							pos1d--
							pos2d[1]--
						}
						if pos2d[1] < 0 {
							pos2d[0]--
							pos2d[1] = len(Split(buffer, "\n")[pos2d[0]]) - lineLen
						}
					}
				}
			} else if key == "enter" {
				if saving {
					err := os.WriteFile(savePath, []byte(buffer), 0644)
					if err != nil {
						clearScreen()
						return err
					}
					clearScreen()
					Println("Saved to", savePath, "\r")
					Print("\x1b[?25h")
					break
				} else {
					buffer = buffer[:pos1d] + "\n" + buffer[pos1d:]
					pos1d++
					pos2d[0]++
					pos2d[1] = 0
					if pos2d[0] >= int(ws.Row)-1 {
						offset++
					}
				}
			} else if key == "esc" {
				if saving {
					saving = false
					Print("\x1b[?25h")
				} else {
					if len(buffer) == 0 {
						clearScreen()
						return nil
					} else {
						for {
							Printf("\x1b[%d;1H%s Do you want to save changes before you exit? (y/n)%s\x1b[0m", int(ws.Row), SELECTEDTEXT, Repeat(" ", int(ws.Col)-51))
							Printf("\x1b[%d;%dH", cursorPos[0], cursorPos[1])
							key, err := readKey()
							if err != nil {
								return err
							}
							if key == "y" {
								saving = true
								break
							} else if key == "n" {
								clearScreen()
								return nil
							} else if key != "" {
								break
							}
						}
					}
				}
			} else if key == "tab" {
				if !saving {
					spaces := 4 - pos2d[1]%4
					buffer = buffer[:pos1d] + Repeat(" ", spaces) + buffer[pos1d:]
					pos1d += spaces
					pos2d[1] += spaces
				}
			} else if isChar(key) {
				if saving {
					if key == " " {
						savePath += "\\ "
					} else {
						savePath += key
					}
				} else {
					buffer = buffer[:pos1d] + key + buffer[pos1d:]
					pos1d++
					pos2d[1]++
				}
			} else if key == "up" {
				if !saving {
					if pos2d[0] > 0 {
						pos2d[0]--
						if pos2d[1] > len(Split(buffer, "\n")[pos2d[0]]) {
							if len(Split(buffer, "\n")[pos2d[0]]) == 0 {
								pos2d[1] = 0
							} else {
								pos2d[1] = len(Split(buffer, "\n")[pos2d[0]]) - 1
							}
						}
						pos1d = 0
						for i := 0; i < pos2d[0]; i++ {
							pos1d += len(Split(buffer, "\n")[i]) + 1
						}
						pos1d += pos2d[1]
					}
					if pos2d[0] < offset {
						offset--
					}
				}
			} else if key == "down" {
				if !saving {
					if pos2d[0] < len(Split(buffer, "\n"))-1 {
						pos2d[0]++
						if pos2d[1] > len(Split(buffer, "\n")[pos2d[0]]) {
							if len(Split(buffer, "\n")[pos2d[0]]) == 0 {
								pos2d[1] = 0
							} else {
								pos2d[1] = len(Split(buffer, "\n")[pos2d[0]]) - 1
							}
						}
						pos1d = 0
						for i := 0; i < pos2d[0]; i++ {
							pos1d += len(Split(buffer, "\n")[i]) + 1
						}
						pos1d += pos2d[1]
					}
					if pos2d[0] >= offset+int(ws.Row)-1 {
						offset++
					}
				}
			} else if key == "right" {
				if !saving {
					if pos2d[1] < len(Split(buffer, "\n")[pos2d[0]]) {
						pos2d[1]++
						pos1d++
					}
				}
			} else if key == "left" {
				if !saving {
					if pos2d[1] > 0 {
						pos2d[1]--
						pos1d--
					}
				}
			} else if isCtrl(key, 's') {
				saving = true
				Print("\x1b[?25l")
			} else if isCtrl(key, 'p') {
				previewMode = !previewMode
				Print("\x1b[?25l")
			} else {
				continue
			}
		}
	}

	return nil
}

func hexToAnsi(hex string, fg bool) string {
	if !HasPrefix(hex, "\"") || !HasSuffix(hex, "\"") {
		Printf("Invalid hex color in config file: %s\n", hex)
		os.Exit(1)
	}
	hex = hex[1 : len(hex)-1]
	if !HasPrefix(hex, "#") {
		Printf("Invalid hex color in config file: %s\n", hex)
		os.Exit(1)
	}
	hex = hex[1:]

	if len(hex) != 6 {
		Printf("Invalid hex color in config file: %s\n", hex)
		os.Exit(1)
	}

	r, err := ParseInt(hex[0:2], 16, 32)
	if err != nil {
		Printf("Invalid hex color in config file: %s (%s)\n", hex, hex[0:2])
		os.Exit(1)
	}

	g, err := ParseInt(hex[2:4], 16, 32)
	if err != nil {
		Printf("Invalid hex color in config file: %s (%s)\n", hex, hex[2:4])
		os.Exit(1)
	}

	b, err := ParseInt(hex[4:6], 16, 32)
	if err != nil {
		Printf("Invalid hex color in config file: %s (%s)\n", hex, hex[4:6])
		os.Exit(1)
	}

	if fg {
		return Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
	} else {
		return Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
	}
}

func parseConfig(contents string) {
	for i := 0; i < len(contents); i++ {
		if contents[i] == '#' {
			for ; i < len(contents) && contents[i] != '\n'; i++ {
			}
		} else if contents[i] == ' ' || contents[i] == '\t' || contents[i] == '\n' {
			continue
		} else {
			var key string
			var value string
			for ; i < len(contents) && contents[i] != ' ' && contents[i] != '\t' && contents[i] != '\n'; i++ {
				key += string(contents[i])
			}
			for ; i < len(contents) && (contents[i] == ' ' || contents[i] == '\t'); i++ {
			}
			for ; i < len(contents) && contents[i] != '\n'; i++ {
				if contents[i] == '"' {
					value += string(contents[i])
					i++
					for ; i < len(contents) && contents[i] != '"'; i++ {
						value += string(contents[i])
					}
					value += string(contents[i])
					break
				}
				value += string(contents[i])
			}
			for ; i < len(contents) && contents[i] != '\n'; i++ {
			}

			if key == "nerd_font" {
				if value == "true" {
					NERD_FONT = true
				} else if value == "false" {
					NERD_FONT = false
				} else {
					Printf("Invalid value for nerd_font in config file: %s\n", value)
					os.Exit(1)
				}
			} else if key == "unicode" {
				if value == "true" {
					UNICODE = true
				} else if value == "false" {
					UNICODE = false
				} else {
					Printf("Invalid value for unicode in config file: %s\n", value)
					os.Exit(1)
				}
			} else if key == "theme" {
				if !HasPrefix(value, "\"") || !HasSuffix(value, "\"") {
					Printf("Invalid theme in config file: %s\n", value)
					os.Exit(1)
				}
				value = value[1 : len(value)-1]
				value = ReplaceAll(value, " ", "-")
				if _, err := os.Stat(THEMES_FOLDER + "/" + value + ".conf"); err != nil {
					Printf("Invalid theme in config file: %s\n", value)
					os.Exit(1)
				}
				contents, err := os.ReadFile(THEMES_FOLDER + "/" + value + ".conf")
				if err != nil {
					Printf("Error reading theme file: %s\n", value)
					os.Exit(1)
				}
				parseConfig(string(contents))
			} else if key == "themes_folder" {
				if !HasPrefix(value, "\"") || !HasSuffix(value, "\"") {
					Printf("Invalid themes_folder in config file: %s\n", value)
					os.Exit(1)
				}
				value = Replace(value[1:len(value)-1], "~", os.Getenv("HOME"), 1)
				THEMES_FOLDER = os.ExpandEnv(value)
			} else if key == "fg_text" {
				TEXT_FG = hexToAnsi(value, true)
			} else if key == "bg_text" {
				TEXT_BG = hexToAnsi(value, false)
			} else if key == "fg_line_num" {
				LINE_NUM_FG = hexToAnsi(value, true)
			} else if key == "bg_line_num" {
				LINE_NUM_BG = hexToAnsi(value, false)
			} else if key == "fg_empty_line" {
				EMPTY_LINE_FG = hexToAnsi(value, true)
			} else if key == "bg_empty_line" {
				EMPTY_LINE_BG = hexToAnsi(value, false)
			} else if key == "fg_status_line" {
				STATUS_LINE_FG = hexToAnsi(value, true)
			} else if key == "bg_status_line" {
				STATUS_LINE_BG = hexToAnsi(value, false)
			} else if key == "fg_selected_num" {
				SELECTED_NUM_FG = hexToAnsi(value, true)
			} else if key == "bg_selected_num" {
				SELECTED_NUM_BG = hexToAnsi(value, false)
			} else if key == "fg_selected_text" {
				SELECTED_TEXT_FG = hexToAnsi(value, true)
			} else if key == "bg_selected_text" {
				SELECTED_TEXT_BG = hexToAnsi(value, false)
			} else if key == "h1" {
				H1 = hexToAnsi(value, true)
			} else if key == "h2" {
				H2 = hexToAnsi(value, true)
			} else if key == "h3" {
				H3 = hexToAnsi(value, true)
			} else if key == "h4" {
				H4 = hexToAnsi(value, true)
			} else if key == "h5" {
				H5 = hexToAnsi(value, true)
			} else if key == "h6" {
				H6 = hexToAnsi(value, true)
			} else {
				Printf("Invalid key in config file: %s\n", key)
				os.Exit(1)
			}
		}
	}

	LINENUM = LINE_NUM_FG + LINE_NUM_BG
	LINETEXT = TEXT_FG + TEXT_BG
	EMPTYLINE = EMPTY_LINE_FG + EMPTY_LINE_BG
	STATUSLINE = STATUS_LINE_FG + STATUS_LINE_BG
	SELECTEDNUM = SELECTED_NUM_FG + SELECTED_NUM_BG
	SELECTEDTEXT = SELECTED_TEXT_FG + SELECTED_TEXT_BG
}

func init() {
	var contents []byte
	if _, err := os.Stat(os.ExpandEnv("$HOME/.config/scratchpad/scratchpad.conf")); err == nil {
		contents, err = os.ReadFile(os.ExpandEnv("$HOME/.config/scratchpad/scratchpad.conf"))
		if err != nil {
			Println(err)
			os.Exit(1)
		}
	} else if _, err := os.Stat(os.ExpandEnv("$HOME/.scratchpad.conf")); err == nil {
		contents, err = os.ReadFile(os.ExpandEnv("$HOME/.scratchpad.conf"))
		if err != nil {
			Println(err)
			os.Exit(1)
		}
	}
	parseConfig(string(contents))
}

func main() {
	oldState, err := setRawTerminal()
	if err != nil {
		Println(err, "\r")
		os.Exit(1)
	}
	defer restoreTerminal(oldState)

	var buffer string
	if len(os.Args) >= 2 {
		if os.Args[1] == "--create-config" {

		} else if os.Args[1] == "-v" || os.Args[1] == "--version" {
			Println("ScratchPad version", VERSION, "\r")
			restoreTerminal(oldState)
			os.Exit(0)
		} else if os.Args[1] == "-h" || os.Args[1] == "--help" {
			Println("Usage: scratchpad [file]\r")
			Println("  -h, --help     display this help and exit\r")
			Println("  -v, --version  output version information and exit\r")
			restoreTerminal(oldState)
			os.Exit(0)
		}
		contents, err := os.ReadFile(os.Args[1])
		if err != nil {
			Println(err, "\r")
			restoreTerminal(oldState)
			os.Exit(1)
		}
		buffer = string(contents)
	}

	err = scratchPad(buffer)
	if err != nil {
		Println(err, "\r")
		restoreTerminal(oldState)
		os.Exit(1)
	}
}
