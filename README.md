# bluetooth-tui2

Bluetooth control TUI in Go using Bubble Tea.

## Run

```bash
go run ./cmd/bluetooth-tui
```

Controls:

- `j` / `k` or arrows: move selection
- `enter`: pair/connect selected device
- `r`: re-scan
- `q`: quit

## Test

```bash
go test ./...
```

Box at top with
j/k or arrows Nav, enter symbol connect, q quit, r rescan,
highlight keys
↑/↓ products


Inspired by primagen's navigation

https://lexfridman.com/theprimeagen-transcript/
```
And so I have this whole style in which I like to navigate, inspired by StarCraft, is that I believe in the press one key, go where you want to be mentality. And so everything about my setup is press one key. So, when I want to go to Twitch chat, alt-two, Twitch chat. When I want to go to my browser, alt-one. That’s my browser. Alt-three, that’s where I go to my programming. That’s power finger, obviously. The big middle finger right there, just smash it down. Alt-six is going to be gimp, so my GNU image manipulation program, so if I want to draw, I go there.
(03:29:02) When I used to have Slack, it was alt-five. If I have a spare terminal where I need to run some extra things, that’s alt-four. I had all these kind of… Everything is perfectly mapped out to single-key. And then when it comes down to using, say, Tmux, I have all my terminals into one single terminal. And now I’m able to kind of switch between there. Prefix one goes to my Vim editor. Whatever project I’m in, it’s always the first Tmux tab, if you will. I’m not sure… They call it a session, but I’m not sure how to describe it if you’re not familiar with Tmux. A tab. Second one is like my spare terminal, third one is my long-running process terminal, my fourth one is a long-running process terminal. So, I have it all set up, so every project I go to automatically spawns session one: Vim, session two: spare terminal, session three will also open it, so it’s like everything’s just ready to rock.
```
