# exslack

Job manager notifying to Slack.   
Slack implemented direct message to myself. this command notify you an end of commands on Slack.


## Usage

1. You have to prepare config file on your home directory `~/.exslackrc` like this.

```json
{
  "webHookURL": "https://hooks.slack.com/services/X..../Y....",
  "destination": "@rompei",
  "logFile": "/home/rompei/joblog.txt"
}
```

- webHookURL: URL of incoming web hook of Slack.
- destination: User name or channel name you want to notify to.
- logFile(optional): If you need output of commands, set this varialbe or set from command line option `-log`.

2. You have to prepare command list file like this.

```
./test1.sh xyz
./test2.sh abc
```

3. Execute this command with command list file.

```bash
./exslack commands.txt
```

4. When the command finish, notify you on channel you specified. And if you set log file, an output will be written the log file.


## Help

```bash
Usage of ./exslack:
  -c  Execute commands concrrentry.
  -cpu uint
      How many CPUs to execution. (default 1)
  -log string
      If you need output of commands, please set this flag or set from config file.
```

## License

[BSD-3](https://opensource.org/licenses/BSD-3-Clause)
