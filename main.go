package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	KeyBackspace = 8
	Tab          = 9
	KeyEnter     = 13
	KeyEscape    = 27

	LINENUM      = "\x1b[38;5;239m\x1b[48;5;234m"
	SELECTEDNUM  = "\x1b[38;5;11m\x1b[48;5;239m"
	EMPTYLINE    = "\x1b[38;5;236m\x1b[48;5;234m"
	LINETEXT     = "\x1b[38;5;15m\x1b[48;5;234m"
	SELECTEDTEXT = "\x1b[38;5;11m\x1b[48;5;239m"
	STATUSLINE   = "\x1b[38;5;15m\x1b[48;5;234m"
	RESET        = "\x1b[0m"

	VERSION = "0.0.1"
)

type Winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func restoreTerminal(oldState *syscall.Termios) {
	fmt.Print("\x1b[?25h")
	if oldState != nil {
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0); err != 0 {
			fmt.Println("Error restoring terminal:", err)
		}
	}
}

func setRawTerminal() (*syscall.Termios, error) {
	fmt.Print("\x1b[?25l")
	oldState := &syscall.Termios{}
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(oldState)), 0, 0, 0); err != 0 {
		return nil, err
	}

	newState := *oldState

	newState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	newState.Iflag &^= syscall.ICRNL | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON

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
		return nil, fmt.Errorf("error getting terminal size: %v", err)
	}
	return ws, nil
}

func length(s string) int {
	return len([]rune(s))
}

func clearScreen() {
	fmt.Print("\x1b[2J")
	fmt.Print("\x1b[3J")
	fmt.Print("\x1b[H")
	os.Stdout.Sync()
}

func scratchPad(buffer string) error {
	var savePath string

	saving := false
	offset := 0
	pos2d := []int{0, 0}
	pos1d := 0

	for {
		var frame []string

		ws, err := getSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}

		intLines := 0
		lines := strings.Split(buffer, "\n")
		numPadding := len(fmt.Sprint(len(lines) + offset))

		for i, line := range lines[offset:] {
			if i == int(ws.Row)-1 {
				break
			}
			var lineNum, lineText string
			if i+offset == pos2d[0] && !saving {
				lineNum = SELECTEDNUM
				lineText = SELECTEDTEXT
			} else {
				lineNum = LINENUM
				lineText = LINETEXT
			}

			addSpaces := 0
			if !saving {
				if i+offset == pos2d[0] {
					if pos2d[1] == len(line) {
						line += "\x1b[7m \x1b[0m" + lineText
						addSpaces = len(lineText) + len("\x1b[7m \x1b[0m") - 1
					} else {
						line = line[:pos2d[1]] + "\x1b[7m" + line[pos2d[1]:pos2d[1]+1] + "\x1b[0m" + lineText + line[pos2d[1]+1:]
						addSpaces = len(lineText) + len("\x1b[7m\x1b[0m") - 1
					}
				}
			}

			if strings.HasPrefix(line, "# ") {
				line = strings.Replace(line, "# ", "\b", 1)
				lineText += "\x1b[1m"
				addSpaces += 2
			} else if strings.HasPrefix(strings.TrimSpace(line), "- [ ] ") || strings.HasPrefix(strings.TrimSpace(line), "> - [ ] ") {
				line = strings.Replace(line, "- [ ] ", "◯ ", 1)
			} else if strings.HasPrefix(strings.TrimSpace(line), "- [x] ") || strings.HasPrefix(strings.TrimSpace(line), "> - [x] ") {
				line = strings.Replace(line, "- [x] ", "⬤ ", 1)
			} else if strings.HasPrefix(strings.TrimSpace(line), "- [X] ") || strings.HasPrefix(strings.TrimSpace(line), "> - [X] ") {
				line = strings.Replace(line, "- [X] ", "⬤ ", 1)
			} else if strings.HasPrefix(strings.TrimSpace(line), "- ") || strings.HasPrefix(strings.TrimSpace(line), "> - ") {
				line = strings.Replace(line, "- ", "• ", 1)
			}
			if strings.HasPrefix(strings.TrimSpace(line), "> ") {
				line = strings.Replace(line, "> ", "▏", 1)
			}

			if len(line) > int(ws.Col)-numPadding-1 {
				frame = append(frame, fmt.Sprintf("%s%s%d%s  %s", lineNum, strings.Repeat(" ", numPadding-len(fmt.Sprint(i+1+offset))), i+1+offset, lineText, line[:int(ws.Col)-numPadding-2]))
				frame = append(frame, fmt.Sprintf("%s  %s%s\x1b[0m", strings.Repeat(" ", numPadding), line[int(ws.Col)-numPadding-1:], strings.Repeat(" ", int(ws.Col)-numPadding-1-length(line[int(ws.Col)-numPadding-2:])+addSpaces)))
				intLines += 2
			} else {
				frame = append(frame, fmt.Sprintf("%s%s%d%s  %s%s\x1b[0m", lineNum, strings.Repeat(" ", numPadding-len(fmt.Sprint(i+1+offset))), i+1+offset, lineText, line, strings.Repeat(" ", int(ws.Col)-length(line)-numPadding-2+addSpaces)))
				intLines += 1
			}
		}

		for i := 0; i < int(ws.Row)-intLines-1; i++ {
			frame = append(frame, fmt.Sprintf("%s~%s\x1b[0m", EMPTYLINE, strings.Repeat(" ", int(ws.Col)-1)))
		}

		if saving {
			frame = append(frame, fmt.Sprintf("%s Save as: %s\x1b[7m \x1b[0m%s%s\x1b[0m", SELECTEDTEXT, savePath, SELECTEDTEXT, strings.Repeat(" ", int(ws.Col)-11-len(savePath))))
		} else {
			frame = append(frame, fmt.Sprintf("%s %d lines %s %d:%d \x1b[0m", STATUSLINE, len(lines), strings.Repeat(" ", int(ws.Col)-11-len(fmt.Sprint(len(lines)))-len(fmt.Sprint(pos2d[0]+1))-len(fmt.Sprint(pos2d[1]+1))), pos2d[0]+1, pos2d[1]+1))
		}

		clearScreen()
		fmt.Print(strings.Join(frame, "\n"))

		key, err := readKey()
		if err != nil {
			return err
		}
		if key == "backspace" {
			if saving {
				if len(savePath) > 0 {
					savePath = savePath[:len(savePath)-1]
				}
			} else {
				if pos1d > 0 {
					lineLen := len(strings.Split(buffer, "\n")[pos2d[0]])
					if pos2d[1] > 3 && strings.HasSuffix(buffer[pos1d-4:pos1d], "    ") && pos2d[1]%4 == 0 {
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
						pos2d[1] = len(strings.Split(buffer, "\n")[pos2d[0]]) - lineLen
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
				fmt.Println("Saved to", savePath)
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
			} else {
				if len(buffer) == 0 {
					clearScreen()
					return nil
				} else {
					for {
						fmt.Printf("\x1b[%d;1H%s Do you want to save changes? (y/n)%s\x1b[0m", int(ws.Row), SELECTEDTEXT, strings.Repeat(" ", int(ws.Col)-35))
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
						} else {
							break
						}
					}
				}
			}
		} else if key == "tab" {
			if !saving {
				spaces := 4 - pos2d[1]%4
				buffer = buffer[:pos1d] + strings.Repeat(" ", spaces) + buffer[pos1d:]
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
					if pos2d[1] > len(strings.Split(buffer, "\n")[pos2d[0]]) {
						if len(strings.Split(buffer, "\n")[pos2d[0]]) == 0 {
							pos2d[1] = 0
						} else {
							pos2d[1] = len(strings.Split(buffer, "\n")[pos2d[0]]) - 1
						}
					}
					pos1d = 0
					for i := 0; i < pos2d[0]; i++ {
						pos1d += len(strings.Split(buffer, "\n")[i]) + 1
					}
					pos1d += pos2d[1]
				}
				if pos2d[0] < offset {
					offset--
				}
			}
		} else if key == "down" {
			if !saving {
				if pos2d[0] < len(strings.Split(buffer, "\n"))-1 {
					pos2d[0]++
					if pos2d[1] > len(strings.Split(buffer, "\n")[pos2d[0]]) {
						if len(strings.Split(buffer, "\n")[pos2d[0]]) == 0 {
							pos2d[1] = 0
						} else {
							pos2d[1] = len(strings.Split(buffer, "\n")[pos2d[0]]) - 1
						}
					}
					pos1d = 0
					for i := 0; i < pos2d[0]; i++ {
						pos1d += len(strings.Split(buffer, "\n")[i]) + 1
					}
					pos1d += pos2d[1]
				}
				if pos2d[0] >= offset+int(ws.Row)-1 {
					offset++
				}
			}
		} else if key == "right" {
			if !saving {
				if pos2d[1] < len(strings.Split(buffer, "\n")[pos2d[0]]) {
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
		}
	}

	return nil
}

func main() {
	oldState, err := setRawTerminal()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer restoreTerminal(oldState)

	var buffer string
	if len(os.Args) == 2 {
		contents, err := os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println(err)
			restoreTerminal(oldState)
			os.Exit(1)
		}
		buffer = string(contents)
	}

	err = scratchPad(buffer)
	if err != nil {
		fmt.Println(err)
		restoreTerminal(oldState)
		os.Exit(1)
	}
}
