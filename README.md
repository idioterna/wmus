# wmus
Music queue web interface using mplayer/cvlc and pafy currently running on a rpi2.
For some reason, vlc makes the audio output a lot less noisy, so we're using that.

Run

```
go run wmus.go :8080
```

Then visit http://host:8080/

![wmus screenshot](/wmus.png "wmus")
