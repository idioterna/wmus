# wmus
Music queue web interface using mplayer/cvlc and pafy currently running on a rpi2.
For some reason, vlc makes the audio output a lot less noisy, so we're using that.

Run

```
$ go run wmus.go :8080
```

Then visit http://host:8080/

![wmus screenshot](/wmus.png "wmus")

## wmus helper userscript
I added a ![wmus helper userscript](/wmus_helper.user.js) script that adds a
simple queue button to youtube watch pages that adds the current url to the
queue. Until it knows where your wmus is running it will display a text input
where you must paste your wmus url.


*NOTE:* Logged in youtube users always get redirected to https URIs and
cross-site requests to non-https URIs are then forbidden (this is a good thing).
You must set up a SSL certificate that your browser trusts for this to work.

*NOTE*: To learn more about SSL security, which this example
is completely devoid of, visit [SSL Labs](https://www.ssllabs.com/)

Instructions for creating a self-signed certificate are below:

```
$ openssl req -x509 -nodes -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3650
```

You may then start your wmus server with TLS support:

```
$ go run wmus.go :8443 cert.pem key.pem
```

and visit https://localhost:8443/ and tell your browser to trust your
self-signed certificate.

You may want to use a https proxy, such as [nginx](http://nginx.org/), with
a configuration similar to this (assuming you're running wmus on :8080):

```
server {
        listen 443;
        server_name localhost;

        ssl on;
        ssl_certificate cert.pem;
        ssl_certificate_key key.pem;

        location / {
                proxy_pass http://localhost:8080;
        }
}
```

