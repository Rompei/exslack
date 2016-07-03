# exslack

[![Build Status](https://drone.io/github.com/Rompei/exslack/status.png)](https://drone.io/github.com/Rompei/exslack/latest)

Job manager notifying to Slack.   
Slack implemented direct message to myself. this command notify you an end of commands on Slack.

[Executable](https://drone.io/github.com/Rompei/exslack/files)

## Usage

1. You have to prepare config file on your home directory `~/.exslackrc` like this.

```json
{
  "webHookURL": "https://hooks.slack.com/services/X..../Y....",
  "destination": "@rompei",
  "logDir": "/home/rompei/log",
  "maxAge": 7,
  "maxBackups": 5,
  "maxSize": 100
}
```

- webHookURL: URL of incoming web hook of Slack.
- destination: User name or channel name you want to notify to.
- logDir(optional): If you need output of commands, set this varialbe or set from command line option `-logfile`.
- maxAge(optional): max age of log file (Default: 7 days).
- maxBackups(optional): The number of log files (Default: 5 files).
- maxSize(optional): Size of a log file (Default: 100MB).

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


5. Also we can execute command instantly

```bash
./exslack echo "aaa"
```

## Help

```bash
Usage of ./exslack:
  -conc
    	Execute commands concrrentry.
  -cpus uint
    	How many CPUs to execution. (default 1)
  -jobs string
    	List of jobs.
  -logfile string
    	If you need output of commands, please set this flag or set directory from config file.
```

## License

[BSD-3](https://opensource.org/licenses/BSD-3-Clause)
