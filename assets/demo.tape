# Sirsi Pantheon Demo — VHS Tape
# Run: vhs assets/demo.tape
# Produces: assets/demo.gif

Output assets/demo.gif

Set Shell "zsh"
Set FontSize 13
Set Width 880
Set Height 480
Set Theme "Catppuccin Mocha"
Set Padding 16
Set TypingSpeed 50ms

Sleep 500ms

Type "sirsi version"
Enter
Sleep 2s

Type "sirsi scan"
Enter
Sleep 10s

Type "sirsi dashboard --no-open"
Enter
Sleep 3s

Screenshot assets/demo.png

Type@30ms ""
Sleep 2s
