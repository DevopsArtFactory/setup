# setup

Setup is a command line tool for assuming role. You can easily manage assume role and change role anytime you want.
<br>
I recommend this tool if you are managing multiple AWS account. Easily get assume credentials and paste it to your terminal.

## Install
* If you want to use setup command, then you need to set up..
 - aws configure

```bash
$ brew tap devopsartfactory/devopsart
$ brew install setup
$ setup version
``` 


## How to use
### First init setup.
* Session name should be your original IAM user name in the account from which you log in through console.
```bash
$ setup init
? Your session name:  <Session name>  
```

### Add new key
```bash
$ setup add
? Key:  dev
? Role ARN:  arn:aws:iam::1234567891011:role/XXXX
```

### Check assume roles
```bash
$ setup list
[current role list]
dev
```

### Use Assume role
- You can choose one key from the list or just specify the key next to command
```bash
$ setup
? Choose account: dev
Assume Credentials copied to clipboard, please paste it.

$ setup dev
Assume Credentials copied to clipboard, please paste it.
```

### Edit role ARN
```bash
$ setup edit
? Choose account to edit: dev
? New role ARN: <New Role ARN>
```

### Delete role
```bash
$ setup delete
? Choose account to delete: dev
dev is deleted

$ setup ls
[current role list]

```
