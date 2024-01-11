# segmago - a sega master system emulator in go

My other emulators:
[dmgo](https://github.com/theinternetftw/dmgo),
[famigo](https://github.com/theinternetftw/famigo),
[vcsgo](https://github.com/theinternetftw/vcsgo), and
[a1go](https://github.com/theinternetftw/a1go).

#### Features:
 * Audio!
 * Saved game support!
 * Quicksave/Quickload, too!
 * Game Gear and VGM file support!
 * Glitches are rare but still totally happen!
 * Graphical and auditory cross-platform support!

#### Dependencies:

 * You can compile on windows with no C dependencies.
 * Other platforms should do whatever the [ebiten](https://github.com/hajimehoshi/ebiten) page says, which is what's currently under the hood.

#### Compile instructions

 * If you have go version >= 1.18, `go build ./cmd/segmago` should be enough.
 * The interested can also see my build script `b` for profiling and such.
 * Non-windows users will need ebiten's dependencies.

#### Important Notes:

 * Keybindings are currently hardcoded to WSAD / JK / TY (arrowpad, ab, start/select)
 * Saved games use/expect a slightly different naming convention than usual: romfilename.(sms or gg).sav
 * Quicksave/Quickload is done by pressing m or l (make or load quicksave), followed by a number key

