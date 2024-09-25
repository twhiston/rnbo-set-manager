# rnbo-set-manager

Manage your rnbo sets on your pi for backup and restore.

If you are using this tool it assumes you are somewhat comfortable poking around
in the rpi and in sqlite databases.

Always backup your database before you do any imports!

## Installation

The easiest way to install this tool is to copy the built binary to your rpi.
Otherwise you will need to install the go toolchain and compile it yourself.


You can curl the latest release from your pi
`curl https://github.com/twhiston/rnbo-set-manager/releases/latest/rnbo-set-manager.zip`
and then unzip it and move it to somewhere in your path
`unzip rnbo-set-manager.zip && sudo mv ./rnbo-set-manager /usr/bin/path`

If you try to export a set and you get the error

```txt
Issue getting set
unable to open database file: out of memory (14)
```

It's actually to do with the permissions of the rnbo folder and you can fix it
as follows:

```bash
chmod 0775 ~/Documents/rnbo
# To undo this you can run
# chmod 0755 ~/Documents/rnbo
```

## Usage

`rnbo-set-manager --help`
shows all possible commands, arguments, flags etc...

DO NOT USE AS sudo USER!
It will affect you home path and won't save the exports where you expect them.

BACKUP YOUR DATABASE BEFORE IMPORTING ANYTHING!!!

## TODO

- dumb db backup command (copy db to a new name in selected folder)
- dumb db restore command
- dumb rnbo folder backup command (copy folder)
