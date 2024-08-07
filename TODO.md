# TODO

* Keyboard navigation
  * [✓] hjkl
  * [✓] Highlight current tile
  * [✓] Enter/backspace -> go down/up?
* Reactive
  * [✓] Progressively build tree and treemap
  * [✓] Resize on SIGWINCH
  * [ ] Listen to fs events for file creation
* UI tweaks
  * [✓] Hide thin tiles
  * [✓] Issue with rightmost columns when thin (e.g. .git)
  * [✓] Send to background with ^Z
  * [✓] Progress indicator (spinner)
  * [ ] Leave last screen in scrollback buffer on exit
  * [✓] Status bar with keyboard commands
  * [ ] Display number of hidden files
* Deletion
  * [ ] Delete selected file/dir
    * [ ] Confirm view
* Zoom in / zoom out
  * [ ] Dynamic view root, control with keyboard
* Treemapping algorithm
  * [ ] Optimize for wide tile aspect ratios?
  * [ ] Add squarify algorithm?
  * Stable algorithm more important if treemap is built progressively
* Error handling
  * [✓] Write errors to stderr in scrollback buffer
  * [ ] Dirs with errors are displayed in red
  * [-] Errors view? -- stderr in scrollback will have to do for now
  * [ ] Show number of errors in statusbar
* Improve handling of opts
  * [ ] Consider using getopt lib
  * [ ] Want: "-a" --> "dux: unrecognized option -a"
  * [ ] Want: "-- -a" --> "dux: cannot access 'afsa': No such file or directory"
* Performance
  * [ ] Profile performance
  * [ ] More pointer usage, less copying of big scructs?
* Misc
  * [ ] Use size on disk by default, not size of contents
  * [ ] Don't cross FS boundaries
  * [ ] "Screenshot" in README
  * [✓] Drop golang/geo dependency
  * [ ] Build macOS, linux binaries in CI
  * [ ] asciinema demo?
  * [ ] man page
