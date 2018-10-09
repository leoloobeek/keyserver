# keyserver

### Compiled Binaries
You can retrieve the latest release of keyserver binaries in the Releases page.

### Build
If you would prefer to build the source yourself, make sure Go 1.10+ is 
installed and execute the following:

```
go get -u github.com/leoloobeek/keyserver
```

This project uses the following dependencies:
- github.com/op/go-logging
- github.com/miekg/dns
- github.com/chzyer/readline

### Usage
Head on over to the wiki for more usage information.

### Contributions
I'm sure there will definitely be bugs, but also this tool was written to match my workflow. If there's something you would find useful feel free to submit an Issue or even a PR!

### HUGE Thanks
Thanks to the following people for their awesome code:
- OJ [@TheColonial](https://twitter.com/TheColonial) as I took most of his DNS code from one of his [live streams](https://www.youtube.com/watch?v=FeH2Yrw68f8)
- [evilsocket](https://twitter.com/evilsocket) for [bettercap](https://github.com/bettercap/bettercap), a really well written Go application which I used as a reference point multiple times, including his readline usage. I almost don't want to mention him here, as my Go code is nowhere near his level and this might look bad on him :D
