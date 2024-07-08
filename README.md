# ScratchPad
- Version: 0.1.3
- Made by: vh8t

# About Scratchpad
> - Lightweight note taking app for terminal written in go
> - Made for fast and simple note taking

# Roadmap
- [x] Simple markdown support
- [x] Line wrap
- [x] Config support
- [x] Themes
- [ ] Full markdown support
- [ ] Mouse support
- [ ] Plugins
- [ ] Extensive controls

# Markdown
## Supported
- You can use all 6 levels of heading by prefixing line with x many hashtags (#) from one to six hashtags
- You can use the tab indent by prefixing line with how many `> ` you want, also behind it can be anything else except heading
- You can use bullets by prefixing line with `- `
- You can use checkboxes by either prefixing line with `- [ ] ` for empty checkbox or with `- [x] ` for checked checkbox

## Unsupported
- In the future I would like to add all the other markdown features that are possible in terminal like bold, underline, italic or strikethrought text

# Controls
ScratchPad has just a couple keybinds but I would like to add more in the future
> - `ctrl+s` Save and exit
> - `ctrl+c` or `esc` Exit without saving
> - `ctrl+p` Toggle preview mode

You can move around with just arrows for now but I would like to add mouse support as well

# Themes
ScratchPad has 9 themes by default, they should be located in `~/.config/scratchpad/themes` by default but it is possible to change the directory if you want
- Dark/Light
- Solarized Dark/Light
- Gruvbox Dark/Light
- Nord
- Tokyo Night

# Config
There are 2 locations where you can store your ScratchPad configuration, first one is `~/.config/scratchpad/scratchpad.conf` and second one is `~/.scratchpad.conf` but the first one is recommended
To configure ScratchPad start with creating the config file in one of the above mentioned places, then you can specify the options

```
# Config might look like this, note that none of the fields are required

themes_folder "~/.config/scratchpad/themes" # This specifies where to look for themes

nerd_font true  # Use nerd font icons instead of ASCII characters for preview mode
unicode   true  # Use unicode icons insteead of ASCII characters for preview mode

theme "tokyo night" # Themes should match the file name in the themes folder but without the file extension, also all spaces are automatically replaced by `-`
                    # theme `tokyo night` would be translated to `$THEMES_FOLDER/tokyo-night.conf`

# Colors
fg_text          "#c0caf5" # Base text color
bg_text          "#1a1b26" # Base background color
fg_line_num      "#565f89" # Line number color
bg_line_num      "#1a1b26" # Line number background color
fg_empty_line    "#565f89" # Empty line text color
bg_empty_line    "#1a1b26" # Empty line background color
fg_status_line   "#a9b1d6" # Status line text color
bg_status_line   "#1a1b26" # Status line background color
fg_selected_num  "#c0caf5" # Selected line number color
bg_selected_num  "#7aa2f7" # Selected line number background color
fg_selected_text "#c0caf5" # Selected line text color
bg_selected_text "#3d59a1" # Selected line background color

# Heading colors
h1 "#7aa2f7"
h2 "#7dcfff"
h3 "#bb9af7"
h4 "#ff9e64"
h5 "#9ece6a"
h6 "#e0af68"
```

File made with scratch-pad