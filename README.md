![screenshot](https://raw.githubusercontent.com/vh8t/scratch-pad/main/screenshot.png)

# ScratchPad
- Version: 0.1.3
- Made by: vh8t

# About Scratchpad
> - Simple movements
> - Markdown like syntax

# Roadmap
- [x] Simple markdown support
- [x] Line wrap
- [ ] Full markdown support
- [ ] Colors

# Controls

```
'# '     Heading 1
'## '    Heading 2
...      Heading 6
'> '     Tab
'- '     Bullet
'- [ ]'  Checkbox
'- [x]'  Checkbox

ctrl+s   Save file as
ctrl+p   Preview mode
esc      Close without save
```

# Configuration
By default ScratchPad only uses ASCII characters for the markdown visuals but you can change it to either unicode icons or even nerd font if you have one installed. You can install nerd font [here](https://www.nerdfonts.com). To configure which one to use you can either make `~/.scratchpad.conf` or `~/.config/scratchpad/scratchpad.conf` file and set the `nerd_font true` or `unicode true` property. Make sure that only one is enabled at the smame time.

File made with scratch-pad