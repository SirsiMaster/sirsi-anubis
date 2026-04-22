# Sirsi Pantheon Demo — VHS Tape
# Run: cd ~/Development/sirsi-pantheon && vhs assets/demo.tape
# Produces: assets/demo.gif

Output assets/demo.gif

Set Shell "zsh"
Set FontSize 13
Set Width 880
Set Height 520
Set Theme "Catppuccin Mocha"
Set Padding 16
Set TypingSpeed 50ms

Sleep 500ms

Type "./bin/sirsi version"
Enter
Sleep 2s

Type "./bin/sirsi scan"
Enter
Sleep 12s

Screenshot assets/demo.png

Sleep 2s
