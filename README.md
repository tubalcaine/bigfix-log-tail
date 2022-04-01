# bigfix-log-tail
A CLI tool, written in Go, which will tail the BigFix agent logs indefinitely. It uses goroutines to write the lines and to find and read the newest file in the client log directory. It assumes the default location for the logs based on the OS, and allows a folder name to be specified on the command line to override it. It will always follow the most recent file in the specified directory.
This really is in no way BigFix specific, but I wrote it to scratch an itch with the way BigFix client logs are written. They do not do it in an "syslog" manner. Usually with syslog, the current log name is fixed and when the log "rolls," the name is changed. The standard "tail" command handles this automatically. But the BigFix agent names the log uniquely each day. So this program sets up two goroutines. One which just writes strings that come in on a chnnel to stdout. Another reads lines from the most recent file in a directory and sends them to the channel, then checks to see if there is a newer file than the current one. If there is, it switches to that file and continues to monitor it.
FIXME: Interstingly, right now, this works on Linux boxes, but is not working on Windows boxes! I do not yet know why. But since it works on Linux (and probably will on Mac), I'm making it public.
You can specify the directory to "watch" on the command line. If you do not, it attempts to default to either:
`C:\Program Files (x86)\BigFix Enterprise\BES Client\__BESData\__Global\Logs`
or
`/var/opt/BESClient/__BESData/__Global/Logs`
This is currently based solely on the value of os.PathSeparator.
So lots of room for improvement!
